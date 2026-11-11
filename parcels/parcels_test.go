package parcels_test

import (
	"database/sql"
	"github.com/Longitude103/ESC-Model/parcels"
	"testing"
)

var p1 = parcels.Parcel{ParcelNo: 1234, AppEff: 0.85,
	Nir:       [12]float64{0, 0, 0, 0, 0.2, 0.4, 0.8, 0.8, 0.5, 0, 0, 0},
	DryEt:     [12]float64{0, 0, 0, 1.1, 0.05, 0.1, 0.2, 0.2, 0.1, 0, 0, 0},
	Et:        [12]float64{0, 0, 0, 1.2, 1.2, 2.5, 4.5, 4.5, 3, 0, 0, 0},
	Pump:      [12]float64{0, 0, 0, 0, 0, 12.3, 21.4, 18.9, 0, 0, 0, 0},
	Ro:        [12]float64{0, 0, 0, 0, 0, 0, 0, 0.35, 0.87, 0, 0, 0},
	Dp:        [12]float64{0, 0, 0, 0, 0, 0, 1.5, 1.1, 1.3, 0, 0, 0},
	SWDel:     [12]float64{0, 0, 0, 0, 1, 13, 15, 14, 5, 0, 0, 0},
	CoeffZone: 2, SoilCode: 622, Area: 40.0, IrrType: sql.NullString{String: "SPRINKLER", Valid: true},
	Crop1:    sql.NullInt64{Int64: 1, Valid: true},
	Crop1Cov: sql.NullFloat64{Float64: 1, Valid: true}, Crop2: sql.NullInt64{Int64: 0, Valid: false},
	Crop2Cov: sql.NullFloat64{Float64: 0, Valid: false}, Crop3: sql.NullInt64{Int64: 0, Valid: false},
	Crop3Cov: sql.NullFloat64{Float64: 0, Valid: false}, Crop4: sql.NullInt64{Int64: 0, Valid: false},
	Crop4Cov: sql.NullFloat64{Float64: 0, Valid: false}, Nrd: "np", Yr: 2014,
	CertNum: sql.NullString{String: "3456", Valid: true}, PointX: 41.4, PointY: 103.0,
	Sw: sql.NullBool{Bool: true, Valid: true}, Gw: sql.NullBool{Bool: true, Valid: true}}

var p2 = parcels.Parcel{ParcelNo: 1236, AppEff: 0.85,
	Nir:       [12]float64{0, 0, 0, 0, 0.2, 0.4, 0.8, 0.8, 0.5, 0, 0, 0},
	DryEt:     [12]float64{0, 0, 0, 0, 0.05, 0.1, 0.2, 0.2, 0.1, 0, 0, 0},
	Et:        [12]float64{0, 0, 0, 0, 1.2, 2.5, 4.5, 4.5, 3, 0, 0, 0},
	SWDel:     [12]float64{0, 0, 0, 0, 0, 12, 15, 14, 5, 0, 0, 0},
	Ro:        [12]float64{0, 0, 0, 0, 0, 0, 0, 0.35, 0.87, 0, 0, 0},
	Dp:        [12]float64{0, 0, 0, 0, 0, 0, 1.5, 1.1, 1.3, 0, 0, 0},
	CoeffZone: 3, Crop1: sql.NullInt64{Int64: 8, Valid: true},
	Crop1Cov: sql.NullFloat64{Float64: 0.5, Valid: true}, Crop2: sql.NullInt64{Int64: 5, Valid: true},
	Crop2Cov: sql.NullFloat64{Float64: 0.5, Valid: true}, Crop3: sql.NullInt64{Int64: 0, Valid: false},
	Crop3Cov: sql.NullFloat64{Float64: 0, Valid: false}, Crop4: sql.NullInt64{Int64: 0, Valid: false},
	Crop4Cov: sql.NullFloat64{Float64: 0, Valid: false}, Nrd: "np", Yr: 2014,
	CertNum: sql.NullString{String: "3456", Valid: true}, PointX: 40.21, PointY: 103.0,
	Sw: sql.NullBool{Bool: true, Valid: true}, Gw: sql.NullBool{Bool: false, Valid: true}}

// p3 is the groundwater only cell made into a parcel from the TFG Example document
var p3 = parcels.Parcel{ParcelNo: 159988, AppEff: 0.65,
	Nir:       [12]float64{0, 0, 0, 0, 0, 0, 4.98, 4.31, 1.65, 0, 0, 0},
	DryEt:     [12]float64{0.24, 0.62, 0.39, 1.36, 1.82, 5.13, 4.55, 2.66, 1.16, 0.70, 0.66, 0.19},
	Et:        [12]float64{0.27, 0.33, 0.82, 1.36, 1.82, 5.13, 7.77, 7.21, 4.02, 0.44, 0.51, 0.23},
	Pump:      [12]float64{0, 0, 0, 0, 0, 0, 2.34, 2.32, 1.4, 0, 0, 0},
	Ro:        [12]float64{0, 0, 0, 1.04, 0.73, 1.81, 0, 0.11, 0, 0.03, 0, 0},
	Dp:        [12]float64{0, 0, 0, 0, 0, 0.39, 0, 0, 0, 0.01, 0, 0},
	CoeffZone: 2, SoilCode: 622, Area: 40.0, IrrType: sql.NullString{String: "FLOOD", Valid: true},
	Crop1: sql.NullInt64{Int64: 1, Valid: true}, Crop1Cov: sql.NullFloat64{Float64: 1, Valid: true},
	Crop2:    sql.NullInt64{Int64: 0, Valid: false},
	Crop2Cov: sql.NullFloat64{Float64: 0, Valid: false}, Crop3: sql.NullInt64{Int64: 0, Valid: false},
	Crop3Cov: sql.NullFloat64{Float64: 0, Valid: false}, Crop4: sql.NullInt64{Int64: 0, Valid: false},
	Crop4Cov: sql.NullFloat64{Float64: 0, Valid: false}, Nrd: "np", Yr: 2014,
	CertNum: sql.NullString{String: "3456", Valid: true}, PointX: 41.4, PointY: 102.5,
	Sw: sql.NullBool{Bool: false, Valid: true}, Gw: sql.NullBool{Bool: true, Valid: true}}

// p3 is the groundwater only cell made into a parcel from the TFG Example document
var p3b = parcels.Parcel{ParcelNo: 159989, AppEff: 0.65,
	Nir:       [12]float64{0, 0, 0, 0, 0, 0, 4.98, 4.31, 1.65, 0, 0, 0},
	DryEt:     [12]float64{0.24, 0.62, 0.39, 1.36, 1.82, 5.13, 4.55, 2.66, 1.16, 0.70, 0.66, 0.19},
	Et:        [12]float64{0.27, 0.33, 0.82, 1.36, 1.82, 5.13, 7.77, 7.21, 4.02, 0.44, 0.51, 0.23},
	Pump:      [12]float64{0, 0, 0, 0, 0, 0, 2.34, 2.32, 1.4, 0, 0, 0},
	Ro:        [12]float64{0, 0, 0, 1.04, 0.73, 1.81, 0, 0.11, 0, 0.03, 0, 0},
	Dp:        [12]float64{0, 0, 0, 0, 0, 0.39, 0, 0, 0, 0.01, 0, 0},
	CoeffZone: 2, SoilCode: 622, Area: 40.0, IrrType: sql.NullString{String: "FLOOD", Valid: true},
	Crop1: sql.NullInt64{Int64: 1, Valid: true}, Crop1Cov: sql.NullFloat64{Float64: 1, Valid: true},
	Crop2:    sql.NullInt64{Int64: 0, Valid: false},
	Crop2Cov: sql.NullFloat64{Float64: 0, Valid: false}, Crop3: sql.NullInt64{Int64: 0, Valid: false},
	Crop3Cov: sql.NullFloat64{Float64: 0, Valid: false}, Crop4: sql.NullInt64{Int64: 0, Valid: false},
	Crop4Cov: sql.NullFloat64{Float64: 0, Valid: false}, Nrd: "np", Yr: 2014,
	CertNum: sql.NullString{String: "3456", Valid: true}, PointX: 41.4, PointY: 102.5,
	Sw: sql.NullBool{Bool: false, Valid: false}, Gw: sql.NullBool{Bool: true, Valid: true}}

// p4 is a parcel with fallow
var p4 = parcels.Parcel{ParcelNo: 1235, AppEff: 0.65,
	Nir:       [12]float64{0, 0, 0, 0, 0, 0, 4.98, 4.31, 1.65, 0, 0, 0},
	DryEt:     [12]float64{0.24, 0.62, 0.39, 1.36, 1.82, 5.13, 4.55, 2.66, 1.16, 0.70, 0.66, 0.19},
	Et:        [12]float64{0.27, 0.33, 0.82, 1.36, 1.82, 5.13, 7.77, 7.21, 4.02, 0.44, 0.51, 0.23},
	Pump:      [12]float64{0, 0, 0, 0, 0, 0, 2.34, 2.32, 1.4, 0, 0, 0},
	Ro:        [12]float64{0, 0, 0, 1.04, 0.73, 1.81, 0, 0.11, 0, 0.03, 0, 0},
	Dp:        [12]float64{0, 0, 0, 0, 0, 0.39, 0, 0, 0, 0.01, 0, 0},
	CoeffZone: 2, SoilCode: 622, Area: 40.0, IrrType: sql.NullString{String: "FLOOD", Valid: true},
	Crop1: sql.NullInt64{Int64: 15, Valid: true}, Crop1Cov: sql.NullFloat64{Float64: 0.5, Valid: true},
	Crop2:    sql.NullInt64{Int64: 8, Valid: false},
	Crop2Cov: sql.NullFloat64{Float64: 0.5, Valid: false}, Crop3: sql.NullInt64{Int64: 0, Valid: false},
	Crop3Cov: sql.NullFloat64{Float64: 0, Valid: false}, Crop4: sql.NullInt64{Int64: 0, Valid: false},
	Crop4Cov: sql.NullFloat64{Float64: 0, Valid: false}, Nrd: "np", Yr: 2014,
	CertNum: sql.NullString{String: "3456", Valid: true}, PointX: 41.4, PointY: 102.5,
	Sw: sql.NullBool{Bool: false, Valid: true}, Gw: sql.NullBool{Bool: true, Valid: true}}

var testParcelSlice = []parcels.Parcel{p1, p2, p3, p4}

func TestParcel_String(t *testing.T) {
	p := parcels.Parcel{ParcelNo: 1234, Nrd: "np", Yr: 2014, Area: 40.0, SoilCode: 622, CoeffZone: 2, AppEff: 0.85,
		IrrType: sql.NullString{String: "SPRINKLER", Valid: true}}
	want := "Parcel No: 1234, NRD: np, Year: 2014, Area: 40.00, SWFac: , SWID: 0, IrrType: SPRINKLER, Soil: 622, Coeff: 2 AppEff: 0.85"
	if got := p.String(); got != want {
		t.Errorf("string doesn't produce correct result:\n got: %q\nwant: %q", got, want)
	}
}

func TestFilterParcelByCert(t *testing.T) {
	sliceP := []parcels.Parcel{p1}
	fp := parcels.FilterParcelByCert(&sliceP, "3456", false, false)

	if sliceP[fp[0]].ParcelNo != 1234 {
		t.Errorf("Didn't return correct parcel go parcel %d instead", sliceP[fp[0]].ParcelNo)
	}
}

func TestFilterParcelByCertNoneFound(t *testing.T) {
	sliceP := []parcels.Parcel{p1}
	fp := parcels.FilterParcelByCert(&sliceP, "6789", false, false)

	if fp != nil {
		t.Error("parcel returned when none should be")
	}
}

func TestParcel_GetXY(t *testing.T) {
	x, y := p1.GetXY()

	if x != 41.4 || y != 103.0 {
		t.Error("not returning the correct X, Y")
	}
}

func TestParcel_changeFallow(t *testing.T) {
	p4.ChangeFallow()

	if p4.Crop1.Int64 != 12 {
		t.Errorf("changeFallow should have changed the crop from fallow (15) to Grass Pasture (12) but got: %d", p4.Crop1.Int64)
	}
}

func TestParcel_noCropCheck(t *testing.T) {
	p := p3 // copy so the shared fixture is not mutated for later tests
	p.Crop1.Int64 = 0
	p.Crop1Cov.Float64 = 0.0

	p.NoCropCheck()

	if p.Crop1.Int64 != 8 || p.Crop1Cov.Float64 != 1.0 {
		t.Errorf("noCropCheck failed, crop should be 8 but got %d, crop coverage should be 1.0 but got %f", p.Crop1.Int64, p.Crop1Cov.Float64)
	}
}

func Test_isGWO(t *testing.T) {
	if p1.IsGWO() {
		t.Error("Parcel 1 is Not Ground water only and returned as it is Groundwater only")
	}

	if !p3.IsGWO() {
		t.Error("Parcel 3 is Ground Water only but returned that is wasn't Groundwater only")
	}

	if !p3b.IsGWO() {
		t.Error("Parcel 3b is Groundwater only but retruned it wasn't GWO, but SW is not valid")
	}
}

func Test_SSParcels(t *testing.T) {
	v := dbConnection(t)
	v.SYear = 1895
	v.EYear = 1905
	v.SteadyState = true

	pcls := parcels.GetParcels(v, 1895)

	if len(pcls) == 0 {
		t.Error("No parcels returned")
	}

	for _, p := range pcls {
		if p.Sw.Bool == false {
			t.Errorf("Parcel has no surface water: %+v\n", p)
		}
	}
}
