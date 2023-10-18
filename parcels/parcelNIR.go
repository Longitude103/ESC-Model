package parcels

import (
	"bitbucket.org/longitude-103/wwumm-esc/database"
	"bitbucket.org/longitude-103/wwumm-esc/fileio"
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
			_, _, _, cropsDryEt[0] = Crop(p.Crop1.Int64, dryData)
		}

		if p.Crop2.Valid {
			cropsNir[1], cropsRo[1], cropsDp[1], cropsEt[1] = Crop(p.Crop2.Int64, annData)
			cropCov[1] = p.Crop2Cov.Float64
			_, _, _, cropsDryEt[1] = Crop(p.Crop2.Int64, dryData)
		}

		if p.Crop3.Valid {
			cropsNir[2], cropsRo[2], cropsDp[2], cropsEt[2] = Crop(p.Crop3.Int64, annData)
			cropCov[2] = p.Crop3Cov.Float64
			_, _, _, cropsDryEt[2] = Crop(p.Crop3.Int64, dryData)
		}

		if p.Crop4.Valid {
			cropsNir[3], cropsRo[3], cropsDp[3], cropsEt[3] = Crop(p.Crop4.Int64, annData)
			cropCov[3] = p.Crop4Cov.Float64
			_, _, _, cropsDryEt[3] = Crop(p.Crop4.Int64, dryData)
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
