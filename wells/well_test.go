package wells

import (
	"database/sql"
	"testing"

	"github.com/Longitude103/ESC-Model/database"
	"github.com/Longitude103/ESC-Model/parcels"
)

var (
	wr1 = database.WelAnnualResult{Wellid: 123, Node: 1, Yr: 2021, FileType: 204, Result: [12]float64{0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0}}
	wr2 = database.WelAnnualResult{Wellid: 124, Node: 1, Yr: 2021, FileType: 204, Result: [12]float64{0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0}}
	wr3 = database.WelAnnualResult{Wellid: 125, Node: 2, Yr: 2021, FileType: 204, Result: [12]float64{0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0}}
	wr4 = database.WelAnnualResult{Wellid: 126, Node: 2, Yr: 2021, FileType: 204, Result: [12]float64{0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0}}

	wrSlice = []database.WelAnnualResult{wr1, wr2, wr3, wr4}

	wn1 = database.WellNode{WellId: 123, RegCd: sql.NullString{String: "1", Valid: true}, Node: 1, Nrd: "np"}
	wn2 = database.WellNode{WellId: 124, RegCd: sql.NullString{String: "2", Valid: true}, Node: 1, Nrd: "np"}
	wn3 = database.WellNode{WellId: 125, RegCd: sql.NullString{String: "3", Valid: true}, Node: 2, Nrd: "sp"}
	wn4 = database.WellNode{WellId: 126, RegCd: sql.NullString{String: "4", Valid: true}, Node: 2, Nrd: "sp"}
	wn5 = database.WellNode{WellId: 127, RegCd: sql.NullString{String: "5", Valid: true}, Node: 2, Nrd: "sp"}

	wnSlice = []database.WellNode{wn1, wn2, wn3, wn4, wn5}

	testParcel1 = parcels.Parcel{ParcelNo: 987, Yr: 2021, Nrd: "np", FirstIrr: sql.NullInt64{Valid: true, Int64: 1999},
		Sw: sql.NullBool{Valid: true, Bool: false}, Pump: [12]float64{0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0}}
	testParcel2 = parcels.Parcel{ParcelNo: 986, Yr: 2021, Nrd: "sp", FirstIrr: sql.NullInt64{Valid: true, Int64: 1972},
		Sw: sql.NullBool{Valid: true, Bool: true}, Pump: [12]float64{0, 0, 0, 0, 2, 2, 2, 2, 2, 0, 0}}

	wp1 = database.WellParcel{WellId: 123, ParcelId: 987, Nrd: "np", Yr: 2021}
	wp2 = database.WellParcel{WellId: 124, ParcelId: 987, Nrd: "np", Yr: 2021}
	wp3 = database.WellParcel{WellId: 125, ParcelId: 986, Nrd: "sp", Yr: 2021}
	wp4 = database.WellParcel{WellId: 126, ParcelId: 986, Nrd: "sp", Yr: 2021}

	wpSlice = []database.WellParcel{wp1, wp2, wp3, wp4}
)

func Test_findResult(t *testing.T) {
	found, location := findResult(wrSlice, 123, 2021, 204)
	if found != true || location != 0 {
		t.Errorf("should have found a result and the location should have been zero; found: %t, location: %d\n", found, location)
	}

	found, location = findResult(wrSlice, 126, 2021, 204)
	if found != true || location != 3 {
		t.Errorf("should have found the well and location; found %t, location %d\n", found, location)
	}

	found, location = findResult(wrSlice, 124, 2020, 204)
	if found == true || location != 0 {
		t.Errorf("should have not found the well and location; found %t, location %d\n", found, location)
	}

	found, location = findResult(wrSlice, 127, 2021, 204)
	if found == true || location != 0 {
		t.Errorf("should have not found the well and location; found %t, location %d\n", found, location)
	}
}

func Test_addToResults(t *testing.T) {
	wrSlice, err := addToResults(wnSlice, wrSlice, 123, testParcel1, 1)
	if err != nil {
		t.Errorf("function errored: %s\n", err)
	}

	if wrSlice[0].Result[4] != 2 {
		t.Errorf("function didn't add to new well result correctly value should be "+
			"Result[4] == 2, values are: %+v\n", wrSlice[0])
	}

	wrSlice, err = addToResults(wnSlice, wrSlice, 127, testParcel2, 2)
	if err != nil {
		t.Errorf("function errored: %s\n", err)
	}

	if wrSlice[len(wrSlice)-1].Wellid != 127 {
		t.Error("new well didn't get appended to slice")
	}
}

func Test_getNode(t *testing.T) {
	nodeNum, err := getNode(wnSlice, 127, "sp")
	if err != nil {
		t.Errorf("function errored: %s\n", err)
	}

	if nodeNum != 2 {
		t.Errorf("node should have returned 2, but returned %d\n", nodeNum)
	}
}

func Test_filterWells(t *testing.T) {
	w, c, err := filterWells(wpSlice, 987, "np", 2021)
	if err != nil {
		t.Errorf("function errored: %s\n", err)
	}

	if c != 2 || w[0] != 123 {
		t.Errorf("well count or id was not correct, should be count 2, got %d, well id should be 123, got %d", c, w[0])
	}

	_, c, err = filterWells(wpSlice, 987, "np", 2020)
	if err != nil {
		t.Errorf("function errored: %s\n", err)
	}

	if c != 0 {
		t.Errorf("well count was not correct, should be count 0, got %d", c)
	}
}
