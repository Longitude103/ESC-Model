package parcels

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Longitude103/ESC-Model/database"
	"github.com/Longitude103/ESC-Model/fileio"
)

// ParcelNIR is a method that adds the NIR, RO, and DP for each parcel from the CSResults and weather station data.
// It produces an intermediate results table of NIR in local sqlite for review and adds three maps to the parcel struct
// Values populated by this method are total acre-feet for the parcel
func (p *Parcel) ParcelNIR(v *database.Setup, Year int, wStations []database.WeatherStation,
	csResults map[string][]fileio.StationResults, it IrrType) error {
	var parcelNIR, parcelRo, parcelDp, parcelEt, parcelDryEt [12]float64

	dist, err := database.Distances(p, wStations)
	if err != nil {
		return err
	}

	if v.AppDebug {
		v.Logger.Infof("Parcel %d; Distances: %+v\n", p.ParcelNo, dist)
	}

	for _, st := range dist {
		var annData, dryData []fileio.StationResults
		for _, data := range csResults[st.Station] {
			if data.Yr == Year && data.Soil == p.SoilCode && data.Irrigation == int(it) {
				annData = append(annData, data)
			}

			if data.Yr == Year && data.Soil == p.SoilCode && data.Irrigation == int(DryLand) {
				dryData = append(dryData, data)
			}
		}

		var cropsNir, cropsRo, cropsDp, cropsEt, cropsDryEt [4][12]float64 // 4 crops X 12 months
		var cropCov [4]float64                                             // crop_coverage
		if p.Crop1.Valid {
			cropsNir[0], cropsRo[0], cropsDp[0], cropsEt[0] = Crop(p.Crop1.Int64, annData)
			cropCov[0] = p.Crop1Cov.Float64
			_, _, _, cropsDryEt[0] = Crop(dryLandCrop(p.Crop1.Int64), dryData)
		}

		if p.Crop2.Valid {
			cropsNir[1], cropsRo[1], cropsDp[1], cropsEt[1] = Crop(p.Crop2.Int64, annData)
			cropCov[1] = p.Crop2Cov.Float64
			_, _, _, cropsDryEt[1] = Crop(dryLandCrop(p.Crop2.Int64), dryData)
		}

		if p.Crop3.Valid {
			cropsNir[2], cropsRo[2], cropsDp[2], cropsEt[2] = Crop(p.Crop3.Int64, annData)
			cropCov[2] = p.Crop3Cov.Float64
			_, _, _, cropsDryEt[2] = Crop(dryLandCrop(p.Crop3.Int64), dryData)
		}

		if p.Crop4.Valid {
			cropsNir[3], cropsRo[3], cropsDp[3], cropsEt[3] = Crop(p.Crop4.Int64, annData)
			cropCov[3] = p.Crop4Cov.Float64
			_, _, _, cropsDryEt[3] = Crop(dryLandCrop(p.Crop4.Int64), dryData)
		}

		// weight the crops based on crop_cov and weather station weight
		parcelNIR = PValues(parcelNIR, cropsNir, cropCov, st.Weight, p.SoilArea)
		parcelEt = PValues(parcelEt, cropsEt, cropCov, st.Weight, p.SoilArea)
		parcelDryEt = PValues(parcelDryEt, cropsDryEt, cropCov, st.Weight, p.SoilArea)
		parcelRo = PValues(parcelRo, cropsRo, cropCov, st.Weight, p.SoilArea)
		parcelDp = PValues(parcelDp, cropsDp, cropCov, st.Weight, p.SoilArea)
	}
	//fmt.Printf("Weighted Parcel ID: %d, NIR is: %v\n", p.ParcelNo, parcelNIR)

	if !v.AppDebug {
		// save parcelNIR to sqlite only for irrigated
		if it == Irrigated {
			err := v.PNirDB.Add(database.PNir{ParcelNo: p.ParcelNo, Nrd: p.Nrd, ParcelNIR: parcelNIR, Year: Year, IrrType: int(it)})
			if err != nil {
				return err
			}
		}
	} else {
		v.Logger.Infof("Weighted Parcel ID: %d, NIR is: %v\n", p.ParcelNo, parcelNIR)
	}

	p.Nir = parcelNIR
	p.Ro = parcelRo
	p.Dp = parcelDp
	p.Et = parcelEt
	p.DryEt = parcelDryEt

	return nil
}

// IrrigatedPasture and Pasture are the CropSim crop codes for irrigated pasture (irrigated only) and pasture (dryland
// only).
const (
	IrrigatedPasture = 12
	Pasture          = 13
)

// dryLandCrop returns the CropSim dryland crop to compare an irrigated crop with in the WSPP ET gain. Irrigated pasture
// has no dryland run in CropSim; its dryland form is pasture.
func dryLandCrop(c int64) int64 {
	if c == IrrigatedPasture {
		return Pasture
	}
	return c
}

// CheckDryLandCrops returns an error if a dryland parcel has irrigated pasture (crop 12) as a crop. CropSim has no
// dryland irrigated pasture, so the parcel would silently get no runoff or deep percolation. Dryland pasture is crop 13.
func (p *Parcel) CheckDryLandCrops() error {
	if slot, ok := p.hasCrop(IrrigatedPasture); ok {
		return fmt.Errorf("dryland parcel has irrigated pasture (crop %d) as crop%d, which CropSim does not have as "+
			"dryland. %s. Change the crop to pasture (crop %d) in the dryland parcel table", IrrigatedPasture, slot,
			p.describe(), Pasture)
	}
	return nil
}

// CheckIrrigatedCrops returns an error if an irrigated parcel has pasture (crop 13) as a crop. CropSim only has pasture
// as dryland, so the parcel would silently get no NIR and no pumping. Irrigated pasture is crop 12.
func (p *Parcel) CheckIrrigatedCrops() error {
	if slot, ok := p.hasCrop(Pasture); ok {
		return fmt.Errorf("irrigated parcel has pasture (crop %d) as crop%d, which CropSim only has as dryland, so it "+
			"would get no NIR or pumping. %s. Change the crop to irrigated pasture (crop %d) in the irrigated parcel table",
			Pasture, slot, p.describe(), IrrigatedPasture)
	}
	return nil
}

// checkCrops runs check on every parcel and returns all the problems found in one error, one parcel per line, so they
// can all be fixed before the next run.
func checkCrops(parcels []Parcel, check func(*Parcel) error) error {
	var errs []error
	for i := range parcels {
		if err := check(&parcels[i]); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// hasCrop returns the first crop slot (1 to 4) of the parcel that is crop c.
func (p *Parcel) hasCrop(c int64) (int, bool) {
	for i, crop := range []sql.NullInt64{p.Crop1, p.Crop2, p.Crop3, p.Crop4} {
		if crop.Valid && crop.Int64 == c {
			return i + 1, true
		}
	}
	return 0, false
}

// describe returns the details a user needs to find the parcel in the database.
func (p *Parcel) describe() string {
	d := fmt.Sprintf("Parcel %d, NRD %s, year %d", p.ParcelNo, strings.ToUpper(p.Nrd), p.Yr)
	if p.CertNum.Valid && p.CertNum.String != "" {
		d += ", cert " + p.CertNum.String
	}
	if p.ModelId.Valid && p.ModelId.String != "" {
		d += ", model id " + p.ModelId.String
	}
	if p.Subarea.Valid && p.Subarea.String != "" {
		d += ", subarea " + p.Subarea.String
	}
	if p.IrrType.Valid && p.IrrType.String != "" {
		d += ", irrigation " + p.IrrType.String
	}
	d += fmt.Sprintf(", %.1f acres", p.Area)
	if p.ParcelNo > 30000 {
		d += " (in a post-97 run, a parcel number over 30000 is a 1997 groundwater-only parcel: subtract 30000 for its id in t1997_irr)"
	}
	return d
}

// Crop function filters the results to the integer crop that is included and returns the NIR, RunOff and Deep Percolation from those
// crops as three arrays.
func Crop(c int64, aData []fileio.StationResults) (nir [12]float64, ro [12]float64, dp [12]float64, et [12]float64) {
	var data fileio.StationResults

	for _, d := range aData {
		if int64(d.Crop) == c {
			data = d
		}
	}

	for i, monthly := range data.MonthlyData {
		nir[i] = monthly.Nir
		ro[i] = monthly.Ro
		dp[i] = monthly.Dp
		et[i] = monthly.Et
	}

	return nir, ro, dp, et
}

// PValues creates the parcel nir, ro, or dp by multiplying the cropNir with crop coverage by station weight to
// return a nir, ro, or dp portion for that parcel.
func PValues(parcelValues [12]float64, cropsNIR [4][12]float64, cropCov [4]float64, stWeight float64, area float64) (values [12]float64) {
	//fmt.Printf("parcelValues: %+v, cropCov: %+v, stWeight: %+v\n", parcelValues, cropCov, stWeight)
	for month, v := range parcelValues {
		values[month] += v
		for crop, cropNir := range cropsNIR {
			values[month] += cropNir[month] * cropCov[crop] * stWeight * area / 12
		}
	}

	return values
}
