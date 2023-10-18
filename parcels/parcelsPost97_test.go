package parcels_test

import (
	"bitbucket.org/longitude-103/wwumm-esc/parcels"
	"database/sql"
	"fmt"
	"testing"
)

// p5 is the groundwater only cell made into a parcel from the TFG Example document
var p5 = parcels.Parcel{ParcelNo: 1237, AppEff: 0.85,
	Nir:       [12]float64{0, 0, 0, 0, 0, 0, 4.98, 4.31, 1.65, 0, 0, 0},
	DryEt:     [12]float64{0.24, 0.62, 0.39, 1.36, 1.82, 5.13, 4.55, 2.66, 1.16, 0.70, 0.66, 0.19},
	Et:        [12]float64{0.27, 0.33, 0.82, 1.36, 1.82, 5.13, 7.77, 7.21, 4.02, 0.44, 0.51, 0.23},
	Pump:      [12]float64{0, 0, 0, 0, 0, 0, 2.34, 2.32, 1.4, 0, 0, 0},
	Ro:        [12]float64{0, 0, 0, 1.04, 0.73, 1.81, 0, 0.11, 0, 0.03, 0, 0},
	Dp:        [12]float64{0, 0, 0, 0, 0, 0.39, 0, 0, 0, 0.01, 0, 0},
	CoeffZone: 2, SoilCode: 622, Area: 40.0, IrrType: sql.NullString{String: "SPRINKLER", Valid: true},
	Crop1: sql.NullInt64{Int64: 1, Valid: true}, Crop1Cov: sql.NullFloat64{Float64: 1, Valid: true},
	Crop2:    sql.NullInt64{Int64: 0, Valid: false},
	Crop2Cov: sql.NullFloat64{Float64: 0, Valid: false}, Crop3: sql.NullInt64{Int64: 0, Valid: false},
	Crop3Cov: sql.NullFloat64{Float64: 0, Valid: false}, Crop4: sql.NullInt64{Int64: 0, Valid: false},
	Crop4Cov: sql.NullFloat64{Float64: 0, Valid: false}, Nrd: "np", Yr: 2014,
	CertNum: sql.NullString{String: "3456", Valid: true}, PointX: 41.4, PointY: 102.5,
	Sw: sql.NullBool{Bool: false, Valid: true}, Gw: sql.NullBool{Bool: true, Valid: true}}

var p97ParcelSlice = []parcels.Parcel{p5, p3b}

func Test_parcelPost97(t *testing.T) {
	p97Parcels := parcels.ParcelsPost97(testParcelSlice, p97ParcelSlice)

	// make sure parcel 159988 isn't in the new slice and that 1237 is in the new slice
	found1237 := false
	for _, parcel := range p97Parcels {
		fmt.Printf("%+v\n\n", parcel)

		if parcel.ParcelNo == 159988 {
			t.Error("Found Parcel 159988 (p3) that should have been removed")
		}

		if parcel.ParcelNo == 1237 {
			found1237 = true
		}
	}

	if !found1237 {
		t.Error("Parcel 1237 wasn't in the new parcel slice")
	}
}

func Test_removeGWO(t *testing.T) {
	pcls := parcels.RemoveGWO(p97ParcelSlice)

	// want no GWO parcels
	for _, parcel := range pcls {
		fmt.Printf("%+v\n\n", parcel)

		if parcel.IsGWO() {
			t.Error("Found a groundwater only parcel and there shouldn't be one")
		}
	}
}

func Test_post97Parcels(t *testing.T) {
	v := dbConnection()
	if err := v.SetYears(1997, 2020); err != nil {
		t.Error("Error Setting years for setup")
	}

	// run post97 first
	v.Post97 = true
	p97 := parcels.GetParcels(v, 2017)

	// change v to be post97 == false and re-run
	v.Post97 = false
	original := parcels.GetParcels(v, 2017)

	p97SWO := 0
	origSWO := 0
	p97Comingled := 0
	origComingled := 0

	for _, parcel := range p97 {
		if parcel.IsSWO() {
			p97SWO += 1
		}

		if parcel.IsComingled() {
			p97Comingled += 1
		}
	}

	for _, parcel := range original {
		if parcel.IsSWO() {
			origSWO += 1
		}

		if parcel.IsComingled() {
			origComingled += 1
		}
	}

	if p97SWO != origSWO {
		t.Errorf("Post 97 SWO parcel count does not equal Original Model SWO parcel count %d != %d", p97SWO, origSWO)
	}

	if p97Comingled != origComingled {
		t.Errorf("Post 97 comingled parcel count does not equal Original Model comingled parcel count %d != %d", p97Comingled, origComingled)
	}
}
