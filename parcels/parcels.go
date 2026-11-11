package parcels

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Longitude103/ESC-Model/database"
	"github.com/Longitude103/ESC-Model/parcels/conveyLoss"
)

type Parcel struct {
	ParcelNo  int             `db:"parcel_id"`
	CertNum   sql.NullString  `db:"cert_num"`
	Area      float64         `db:"area"`
	Crop1     sql.NullInt64   `db:"crop1"`
	Crop2     sql.NullInt64   `db:"crop2"`
	Crop3     sql.NullInt64   `db:"crop3"`
	Crop4     sql.NullInt64   `db:"crop4"`
	IrrType   sql.NullString  `db:"irrig_type"`
	FirstIrr  sql.NullInt64   `db:"first_irr"`
	SwFac     sql.NullString  `db:"sw_fac"`
	ModelId   sql.NullString  `db:"model_id"`
	Crop1Cov  sql.NullFloat64 `db:"crop1_cov"`
	Crop2Cov  sql.NullFloat64 `db:"crop2_cov"`
	Crop3Cov  sql.NullFloat64 `db:"crop3_cov"`
	Crop4Cov  sql.NullFloat64 `db:"crop4_cov"`
	Sw        sql.NullBool    `db:"sw"`
	Gw        sql.NullBool    `db:"gw"`
	Subarea   sql.NullString  `db:"subarea"`
	Oa        sql.NullInt64   `db:"oa"`
	Nrd       string          `db:"nrd"`
	SwID      sql.NullInt64   `db:"sw_id"`
	PointX    float64         `db:"pointx"`
	PointY    float64         `db:"pointy"`
	SoilArea  float64         `db:"s_area"`
	SoilCode  int             `db:"soil_code"`
	CoeffZone int             `db:"coeff_zone"`
	Yr        int
	AppEff    float64
	Et        [12]float64
	DryEt     [12]float64
	Nir       [12]float64
	Ro        [12]float64
	Dp        [12]float64
	Pump      [12]float64
	SWDel     [12]float64
}

type IrrType int

const (
	Irrigated IrrType = 3
	DryLand   IrrType = 1
)

// GetParcels returns a list of all parcels with crops irrigation types and areas. Returns data for both nrds. There
// can be multiples of the same parcels listed with different soil types. It sets the year into a field in the struct.
func GetParcels(v *database.Setup, Year int) []Parcel {
	if v.SteadyState {
		return getSSParcels(v, Year)
	}

	query := fmt.Sprintf(`SELECT parcel_id, a.crop_int crop1, crop1_cov, b.crop_int crop2, crop2_cov, c.crop_int crop3, crop3_cov, d.crop_int crop4, crop4_cov, sw, gw, subarea, oa,
	irrig_type, sw_fac, first_irr, cert_num::varchar, model_id, sw_id, st_area(i.geom)/43560 area, 'np' nrd,
	st_x(st_transform(st_centroid(i.geom), 4326)) pointx, st_y(st_transform(st_centroid(i.geom), 4326)) pointy,
	sum(st_area(st_intersection(m.geom, i.geom))/43560) s_area, m.soil_code, m.coeff_zone
FROM np.t%d_irr i inner join public.model_cells m on st_intersects(i.geom, m.geom)
				 LEFT join public.crops a on crop1 = a.crop_name
				 LEFT join public.crops b on crop2 = b.crop_name
				 LEFT join public.crops c on crop3 = c.crop_name
				 LEFT join public.crops d on crop4 = d.crop_name
where m.cell_type = %d
GROUP BY parcel_id, a.crop_int, parcel_id, crop1_cov, b.crop_int, crop2_cov, c.crop_int, crop3_cov, d.crop_int, crop4_cov, sw, gw, irrig_type, sw_fac, cert_num::varchar, model_id, st_area(i.geom)/43560, st_x(st_transform(st_centroid(i.geom), 4326)), st_y(st_transform(st_centroid(i.geom), 4326)), m.soil_code, crop1_cov, crop2, crop2_cov, crop3, crop3_cov, crop4, crop4_cov, sw, gw, subarea, oa, irrig_type,
	  sw_fac, first_irr, cert_num::varchar, model_id, st_area(i.geom)/43560, st_x(st_transform(st_centroid(i.geom), 4326)),
	  st_y(st_transform(st_centroid(i.geom), 4326)), nrd, m.soil_code, m.coeff_zone
UNION ALL
SELECT parcel_id, a.crop_int crop1, crop1_cov, b.crop_int crop2, crop2_cov, c.crop_int crop3, crop3_cov, d.crop_int crop4, crop4_cov, sw, gw, subarea, 0 oa,
	irr_type as irrig_type, sw_fac, first_irr, i.id as cert_num, null as model_id, sw_id, st_area(i.geom)/43560 area, 'sp' nrd,
	st_x(st_transform(st_centroid(i.geom), 4326)) pointx, st_y(st_transform(st_centroid(i.geom), 4326)) pointy,
	sum(st_area(st_intersection(m.geom, i.geom))/43560) s_area, m.soil_code, m.coeff_zone
FROM sp.t%d_irr i inner join public.model_cells m on st_intersects(i.geom, m.geom)
				 LEFT join public.crops a on crop1 = a.crop_name
				 LEFT join public.crops b on crop2 = b.crop_name
				 LEFT join public.crops c on crop3 = c.crop_name
				 LEFT join public.crops d on crop4 = d.crop_name
where m.cell_type = %d
GROUP BY parcel_id, a.crop_int, parcel_id, crop1_cov, b.crop_int, crop2_cov, c.crop_int, crop3_cov, d.crop_int, crop4_cov, sw, gw, subarea, oa, irrig_type,
	  sw_fac, first_irr, i.id, model_id, st_area(i.geom)/43560, st_x(st_transform(st_centroid(i.geom), 4326)),
	  st_y(st_transform(st_centroid(i.geom), 4326)), nrd, m.soil_code, m.coeff_zone;`,
		Year, v.CellType(), Year, v.CellType())

	var parcels []Parcel
	err := v.PgDb.Select(&parcels, query)
	if err != nil {
		v.Logger.Errorf("Error in getting parcels for year %d, error: %s", Year, err)
	}

	// remove parcels that are gw == true and sw == false if v.TotalDep is true
	if v.TotalDep {
		for i := 0; i < len(parcels); i++ {
			if parcels[i].IsGWO() {
				parcels = append(parcels[:i], parcels[i+1:]...)
				i--
			}
		}
	}

	for i := 0; i < len(parcels); i++ {
		parcels[i].Yr = Year
		parcels[i].ChangeFallow()
		parcels[i].NoCropCheck()
	}

	if Year > 1997 && v.Post97 {
		p97GWO := get97GWOParcels(v, Year)
		parcels = ParcelsPost97(parcels, p97GWO)
	}

	return parcels
}

// get97GWOParcels returns a list of groundwater only parcels with crops irrigation types and areas. Returns data for both nrds. There
// can be multiples of the same parcels listed with different soil types. It sets the year into a field in the struct.
func get97GWOParcels(v *database.Setup, Year int) []Parcel {
	query := fmt.Sprintf(`SELECT parcel_id, a.crop_int crop1, crop1_cov, b.crop_int crop2, crop2_cov, c.crop_int crop3, crop3_cov, d.crop_int crop4, crop4_cov, sw, gw, subarea, oa, irrig_type, sw_fac, first_irr, cert_num::varchar, model_id, sw_id, st_area(i.geom)/43560 area, 'np' nrd, st_x(st_transform(st_centroid(i.geom), 4326)) pointx, st_y(st_transform(st_centroid(i.geom), 4326)) pointy, sum(st_area(st_intersection(m.geom, i.geom))/43560) s_area, m.soil_code, m.coeff_zone
	FROM np.t1997_irr i
		inner join (select geom, node, soil_code, zone, coeff_zone, mtg, nat_veg from model_cells where cell_type = %d) m on st_intersects(i.geom, m.geom)
		LEFT join public.crops a on crop1 = a.crop_name
		LEFT join public.crops b on crop2 = b.crop_name
		LEFT join public.crops c on crop3 = c.crop_name
		LEFT join public.crops d on crop4 = d.crop_name
	WHERE sw = false and gw = true
	GROUP BY parcel_id, a.crop_int, parcel_id, crop1_cov, b.crop_int, crop2_cov, c.crop_int, crop3_cov, d.crop_int, crop4_cov, sw, gw, irrig_type, sw_fac, cert_num::varchar, model_id, st_area(i.geom)/43560, st_x(st_transform(st_centroid(i.geom), 4326)), st_y(st_transform(st_centroid(i.geom), 4326)), m.soil_code, crop1_cov, crop2, crop2_cov, crop3, crop3_cov, crop4, crop4_cov, sw, gw, irrig_type, sw_fac, first_irr, cert_num::varchar, model_id, st_area(i.geom)/43560, st_x(st_transform(st_centroid(i.geom), 4326)), st_y(st_transform(st_centroid(i.geom), 4326)), nrd, m.soil_code, m.coeff_zone
	UNION ALL
	SELECT parcel_id, a.crop_int crop1, crop1_cov, b.crop_int crop2, crop2_cov, c.crop_int crop3, crop3_cov, d.crop_int crop4, crop4_cov, sw, gw, subarea, 0 oa, irr_type as irrig_type, sw_fac, first_irr, i.id as cert_num, null as model_id, sw_id, st_area(i.geom)/43560 area, 'sp' nrd, st_x(st_transform(st_centroid(i.geom), 4326)) pointx, st_y(st_transform(st_centroid(i.geom), 4326)) pointy, sum(st_area(st_intersection(m.geom, i.geom))/43560) s_area, m.soil_code, m.coeff_zone
	FROM sp.t1997_irr i
		inner join (select geom, node, soil_code, zone, coeff_zone, mtg, nat_veg from model_cells where cell_type = %d) m on st_intersects(i.geom, m.geom)
		LEFT join public.crops a on crop1 = a.crop_name
		LEFT join public.crops b on crop2 = b.crop_name
		LEFT join public.crops c on crop3 = c.crop_name
		LEFT join public.crops d on crop4 = d.crop_name
	WHERE sw = false and gw = true
	GROUP BY parcel_id, a.crop_int, parcel_id, crop1_cov, b.crop_int, crop2_cov, c.crop_int, crop3_cov, d.crop_int, crop4_cov, sw, gw, irrig_type, sw_fac, first_irr, i.id, model_id, st_area(i.geom)/43560, st_x(st_transform(st_centroid(i.geom), 4326)), st_y(st_transform(st_centroid(i.geom), 4326)), nrd, m.soil_code, m.coeff_zone;`, v.CellType(), v.CellType())

	var parcels []Parcel
	err := v.PgDb.Select(&parcels, query)
	if err != nil {
		v.Logger.Errorf("Error in getting parcels for year %d, error: %s", Year, err)
	}

	for i := 0; i < len(parcels); i++ {
		parcels[i].Yr = Year
		parcels[i].ParcelNo = parcels[i].ParcelNo + 30000 // was having duplicates, now will be unique
		parcels[i].ChangeFallow()
		parcels[i].NoCropCheck()
	}

	return parcels
}

// FilterParcelByCert filters a slice of parcels by the CertNum and returns a slice of the parcels that have that CertNum.
func FilterParcelByCert(p *[]Parcel, c string, filterGWO, filterP97GWO bool) (fParcels []int) {
	for i := 0; i < len(*p); i++ {
		if filterGWO && (*p)[i].IsGWO() {
			continue
		}

		if filterP97GWO {
			if (*p)[i].IsPost97() && (*p)[i].IsGWO() {
				continue
			}
		}

		if (*p)[i].CertNum.String == c {
			fParcels = append(fParcels, i)
		}
	}

	return fParcels
}

// parcelSWDelivery method uses the diversions to then calculate the total amount of surface water delivered to a parcel
// from those diversions. It returns nothing, but sets SWDel inside the Parcel
func (p *Parcel) parcelSWDelivery(diversions []conveyLoss.Diversion) {
	canalDivs := conveyLoss.FilterDivs(diversions, int(p.SwID.Int64))

	var swDelivery [12]float64
	for m := 0; m < 12; m++ {
		for _, d := range canalDivs {
			if int(d.DivDate.Time.Month()) == m+1 {
				swDelivery[m] = d.DivAmount.Float64 * p.Area
			}
		}
	}

	p.SWDel = swDelivery
}

// GetDryParcels is a function that returns a slice of Parcel for a year of the dryland only parcels in the model.
func GetDryParcels(v *database.Setup, Year int) []Parcel {
	if v.SteadyState {
		return getDrySSParcels(v, Year)
	}

	query := fmt.Sprintf(`SELECT i.parcel_id, a.crop_int crop1, crop1_cov, b.crop_int crop2, crop2_cov, c.crop_int crop3, crop3_cov, d.crop_int crop4, crop4_cov,
       st_area(i.geom)/43560 area, 'np' nrd, st_x(st_transform(st_centroid(i.geom), 4326)) pointx,
       st_y(st_transform(st_centroid(i.geom), 4326)) pointy, sum(st_area(st_intersection(m.geom, i.geom))/43560) s_area,
       m.soil_code, m.coeff_zone
FROM np.t%d_dry i inner join (select geom, node, soil_code, zone, coeff_zone, mtg, nat_veg from model_cells where cell_type = %d) m on st_intersects(i.geom, m.geom)
    LEFT join public.crops a on crop1 = a.crop_name
    LEFT join public.crops b on crop2 = b.crop_name
    LEFT join public.crops c on crop3 = c.crop_name
    LEFT join public.crops d on crop4 = d.crop_name
GROUP BY i.parcel_id, a.crop_int, parcel_id, crop1_cov, b.crop_int, crop2_cov, c.crop_int, crop3_cov, d.crop_int, crop4_cov,
    st_area(i.geom)/43560, st_x(st_transform(st_centroid(i.geom), 4326)), st_y(st_transform(st_centroid(i.geom), 4326)),
    m.soil_code, crop1_cov, crop2, crop2_cov, crop3, crop3_cov, crop4, crop4_cov, st_area(i.geom)/43560,
    st_x(st_transform(st_centroid(i.geom), 4326)), st_y(st_transform(st_centroid(i.geom), 4326)), nrd, m.soil_code, m.coeff_zone
UNION ALL
SELECT i.parcel_id, a.crop_int crop1, crop1_cov, b.crop_int crop2, crop2_cov, c.crop_int crop3, crop3_cov, d.crop_int crop4, crop4_cov,
       st_area(i.geom)/43560 area, 'sp' nrd, st_x(st_transform(st_centroid(i.geom), 4326)) pointx,
       st_y(st_transform(st_centroid(i.geom), 4326)) pointy, sum(st_area(st_intersection(m.geom, i.geom))/43560) s_area,
       m.soil_code, m.coeff_zone
FROM sp.t%d_dry i inner join (select geom, node, soil_code, zone, coeff_zone, mtg, nat_veg from model_cells where cell_type = %d) m on st_intersects(i.geom, m.geom)
    LEFT join public.crops a on crop1 = a.crop_name
    LEFT join public.crops b on crop2 = b.crop_name
    LEFT join public.crops c on crop3 = c.crop_name
    LEFT join public.crops d on crop4 = d.crop_name
GROUP BY i.parcel_id, a.crop_int, parcel_id, crop1_cov, b.crop_int, crop2_cov, c.crop_int, crop3_cov, d.crop_int, crop4_cov,
    st_area(i.geom)/43560, st_x(st_transform(st_centroid(i.geom), 4326)), st_y(st_transform(st_centroid(i.geom), 4326)),
    m.soil_code, crop1_cov, crop2, crop2_cov, crop3, crop3_cov, crop4, crop4_cov, st_area(i.geom)/43560,
    st_x(st_transform(st_centroid(i.geom), 4326)), st_y(st_transform(st_centroid(i.geom), 4326)), nrd, m.soil_code, m.coeff_zone;`, Year, v.CellType(), Year, v.CellType())

	var parcels []Parcel
	err := v.PgDb.Select(&parcels, query)
	if err != nil {
		v.Logger.Errorf("Error in getting dryland parcels for year %d", Year)
	}

	for i := 0; i < len(parcels); i++ {
		parcels[i].Yr = Year
	}

	if v.AppDebug {
		return parcels[:20]
	}

	return parcels
}

// String is a method of the parcel to return a string of data about the parcel for identification
func (p *Parcel) String() string {
	return fmt.Sprintf("Parcel No: %d, NRD: %s, Year: %d, Area: %.2f, SWFac: %s, SWID: %d, IrrType: %s, Soil: %d, Coeff: %d AppEff: %.2f", p.ParcelNo, p.Nrd, p.Yr, p.Area, p.SwFac.String, p.SwID.Int64, p.IrrType.String, p.SoilCode, p.CoeffZone, p.AppEff)
}

// NIRString is a method to return the string of the NIR values and parcel number
func (p *Parcel) NIRString() string {
	var nirString string
	for i, n := range p.Nir {
		nirString += strconv.FormatFloat(n, 'f', 2, 64)

		if i < 11 {
			nirString += ", "
		}
	}

	return fmt.Sprintf("Parcel No: %d, NIR (acre-feet): %s", p.ParcelNo, nirString)
}

// SWString is a method to return a string of data of the surface water delivery and canal id of a parcel
func (p *Parcel) SWString() string {
	var swString string
	for i, n := range p.SWDel {
		swString += strconv.FormatFloat(n, 'f', 2, 64)

		if i < 11 {
			swString += ", "
		}
	}

	return fmt.Sprintf("Parcel No: %d, SWID: %d, SWDel (acre-feet): %s", p.ParcelNo, p.SwID.Int64, swString)
}

// RoString is a method to return a string of formatted data of the runoff data of a parcel
func (p *Parcel) RoString() string {
	var roString string
	for i, n := range p.Ro {
		roString += strconv.FormatFloat(n, 'f', 2, 64)

		if i < 11 {
			roString += ", "
		}
	}

	return fmt.Sprintf("Parcel No: %d, RunOff (acre-feet): %s", p.ParcelNo, roString)
}

// DpString is a method to return a string of formatted data of the deep percolation data of a parcel
func (p *Parcel) DpString() string {
	var dpString string
	for i, n := range p.Dp {
		dpString += strconv.FormatFloat(n, 'f', 2, 64)

		if i < 11 {
			dpString += ", "
		}
	}

	return fmt.Sprintf("Parcel No: %d, DeepPerc (acre-feet): %s", p.ParcelNo, dpString)
}

// PumpString is a method to return a string of data of the pumping
func (p *Parcel) PumpString() string {
	var pString string
	for i, n := range p.Pump {
		pString += strconv.FormatFloat(n, 'f', 2, 64)

		if i < 11 {
			pString += ", "
		}
	}

	return fmt.Sprintf("Parcel No: %d, Pumping (acre-feet): %s", p.ParcelNo, pString)
}

// PrintNIR is a method to return a string of the NIR of the parcel into a readable format
func (p *Parcel) PrintNIR() string {
	str := strings.Builder{}

	for m, v := range p.Nir {
		str.WriteString(time.Month(m + 1).String())
		str.WriteString(": ")
		str.WriteString(strconv.FormatFloat(v, 'f', 2, 64))
		str.WriteRune('\n')
	}

	return str.String()
}

// GetXY is a method that returns the x and y coordinates of the centroid of the parcel
func (p *Parcel) GetXY() (x float64, y float64) {
	return p.PointX, p.PointY
}

// SetWelFileType is a method that returns the file type of the well that will be assigned pumping.
func (p *Parcel) SetWelFileType() (fileType int, err error) {
	if p.Nrd == "np" {
		if p.FirstIrr.Int64 < 1998 || p.Yr < 1998 {
			// pre1998 condition
			if p.Sw.Bool {
				return 201, nil
			} else {
				return 202, nil
			}
		} else {
			// post 97 parcel
			if p.Sw.Bool {
				return 203, nil
			} else {
				return 204, nil
			}
		}
	}

	if p.Nrd == "sp" {
		if p.FirstIrr.Int64 < 1998 || p.Yr < 1998 {
			// pre1998 condition
			if p.Sw.Bool {
				return 205, nil
			} else {
				return 206, nil
			}
		} else {
			// post 97 parcel
			if p.Sw.Bool {
				return 207, nil
			} else {
				return 208, nil
			}
		}
	}

	fmt.Printf("Parcel data %+v\n", p)
	return 0, errors.New("could not determine file type")
}

// ChangeFallow is a method that changes any parcel with fallow to winter wheat as fallow is already built into winter wheat
// and rotates in that data.
func (p *Parcel) ChangeFallow() {
	if p.Crop1.Int64 == 15 {
		p.Crop1.Int64 = 12
	}

	if p.Crop2.Int64 == 15 {
		p.Crop2.Int64 = 12
	}

	if p.Crop3.Int64 == 15 {
		p.Crop3.Int64 = 12
	}

	if p.Crop4.Int64 == 15 {
		p.Crop4.Int64 = 12
	}
}

// NoCropCheck is a method to ensure that the parcel includes a crop to prevent errors in subsequent processes. It defaults
// a parcel to all corn if there is no crop present
func (p *Parcel) NoCropCheck() {
	cropT := p.Crop1.Int64 + p.Crop2.Int64 + p.Crop3.Int64 + p.Crop4.Int64

	if cropT == 0 {
		p.Crop1.Int64 = 8
		p.Crop1.Valid = true
		p.Crop1Cov.Float64 = 1.0
		p.Crop1Cov.Valid = true
	}
}

// IsGWO is a method that returns a bool if the parcel is groundwater only
func (p *Parcel) IsGWO() bool {
	if p.IsGW() {
		if !p.Sw.Valid || !p.Sw.Bool {
			return true
		}
	}

	return false
}

// IsGW is a method that returns a bool if the parcel receives supply from groundwater
func (p *Parcel) IsGW() bool {
	if p.Gw.Valid && p.Gw.Bool {
		return true
	}

	return false
}

// IsSW is a method that returns a bool if the parcel receives surface water
func (p *Parcel) IsSW() bool {
	if p.Sw.Valid && p.Sw.Bool {
		return true
	}

	return false
}

// IsSWO is a method that returns a bool true if a parcel is surface water only
func (p *Parcel) IsSWO() bool {
	if !p.IsGW() {
		if p.Sw.Valid && p.Sw.Bool {
			return true
		}
	}

	return false
}

func (p *Parcel) IsComingled() bool {
	if p.IsGW() {
		if p.IsSW() {
			return true
		}
	}

	return false
}

func (p *Parcel) IsPost97() bool {
	if p.FirstIrr.Valid {
		if p.FirstIrr.Int64 > 1997 {
			return true
		}
	}

	return false
}

// GetFileType is a method to get the Filetype int for parcel.
func (p *Parcel) GetFileType() (int, error) {
	switch {
	case p.IsComingled(): // co-mingled
		if p.Nrd == "sp" {
			if p.IsPost97() {
				return 207, nil
			}
			return 205, nil
		}
		if p.IsPost97() {
			return 203, nil
		}
		return 201, nil
	case p.IsSWO(): // SW Only
		if p.Nrd == "sp" {
			return 104, nil
		}
		return 103, nil
	case p.IsGWO(): // GW Only
		if p.Nrd == "sp" {
			if p.IsPost97() {
				return 208, nil
			}
			return 206, nil
		}
		if p.IsPost97() {
			return 204, nil
		}
		return 202, nil
	default:
		return 200, errors.New("cannot classify parcel")
	}
}
