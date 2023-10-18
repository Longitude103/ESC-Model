package database_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"bitbucket.org/longitude-103/wwumm-esc/Utils"
	"bitbucket.org/longitude-103/wwumm-esc/database"
	"github.com/joho/godotenv"
)

func getTestDBConn() (*database.Setup, error) {
	var myEnv map[string]string
	myEnv, err := godotenv.Read("../.env")
	if err != nil {
		fmt.Println("Cannot load Env Variables:", err)
		os.Exit(1)
	}

	var v *database.Setup

	var opts []database.Option
	opts = append(opts, database.WithSteadyState(1893, 1952))
	opts = append(opts, database.WithNoSQLite())

	v, err = database.NewSetup(myEnv, opts...)
	if err != nil {
		return v, fmt.Errorf("could not setup database connection: %s", err)
	}

	return v, nil
}

func TestGetSSCellAreas1(t *testing.T) {
	v, err := getTestDBConn()
	if err != nil {
		t.Error("error in database setup", err)
	}

	cells, err := database.GetSSCellAreas1(v)
	if err != nil {
		t.Error("error in getting cells from db", err)
	}

	want := database.CellIntersect{
		Node:     142932,
		Soil:     622,
		CZone:    5,
		CellArea: 40,
	}

	var got database.CellIntersect
	for _, c := range cells {
		if c.Node == 142932 {
			got = c
		}
	}

	if want.Soil != got.Soil {
		t.Errorf("wanted soil %d but got soil %d", want.Soil, got.Soil)
	}

	if want.CZone != got.CZone {
		t.Errorf("wanted CZone %d but got CZone %d", want.CZone, got.CZone)
	}

	if want.CellArea != got.CellArea {
		t.Errorf("wanted CellArea %f but got CellArea %f", want.CellArea, got.CellArea)
	}
}

func TestGetSSCellAreas2(t *testing.T) {
	v, err := getTestDBConn()
	if err != nil {
		t.Error("error in database setup", err)
	}

	cells, err := database.GetSSCellAreas2(v)
	if err != nil {
		t.Error("error in getting cells from db", err)
	}

	wantNP := database.CellIntersect{
		Node:      4294,
		Soil:      722,
		CZone:     1,
		CellArea:  10,
		NpIrrArea: sql.NullFloat64{Float64: 1.53, Valid: true},
		NpDryArea: sql.NullFloat64{Float64: 1.65, Valid: true},
	}

	wantSP := database.CellIntersect{
		Node:      106645,
		Soil:      622,
		CZone:     5,
		CellArea:  40,
		SpIrrArea: sql.NullFloat64{Float64: 10.29, Valid: true},
		SpDryArea: sql.NullFloat64{Float64: 15.24, Valid: true},
	}

	var gotNP database.CellIntersect
	for _, c := range cells {
		if c.Node == 4294 {
			gotNP = c
			break
		}
	}

	if wantNP.Soil != gotNP.Soil {
		t.Errorf("didn't recieve correct soils data for NP test parcel, wanted %d, but got %d", wantNP.Soil, gotNP.Soil)
	}

	if wantNP.CZone != gotNP.CZone {
		t.Errorf("wanted CZone %d but got CZone %d", wantNP.CZone, gotNP.CZone)
	}

	if wantNP.CellArea != gotNP.CellArea {
		t.Errorf("wanted CellArea %f but got CellArea %f", wantNP.CellArea, gotNP.CellArea)
	}

	if wantNP.NpIrrArea.Float64 != Utils.RoundTo(gotNP.NpIrrArea.Float64, 2) {
		t.Errorf("wanted NP Irrigated area of %f but got %f", wantNP.NpIrrArea.Float64, Utils.RoundTo(gotNP.NpIrrArea.Float64, 2))
	}

	if wantNP.NpDryArea.Float64 != Utils.RoundTo(gotNP.NpDryArea.Float64, 2) {
		t.Errorf("wanted NP Dry area of %f but got %f", wantNP.NpDryArea.Float64, Utils.RoundTo(gotNP.NpDryArea.Float64, 2))
	}

	var gotSP database.CellIntersect
	for _, c := range cells {
		if c.Node == 106645 {
			gotSP = c
			break
		}
	}

	if wantSP.Soil != gotSP.Soil {
		t.Errorf("didn't recieve correct soils data for SP test parcel, wanted %d, but got %d", wantSP.Soil, gotSP.Soil)
	}

	if wantSP.CZone != gotSP.CZone {
		t.Errorf("wanted CZone %d but got CZone %d", wantSP.CZone, gotSP.CZone)
	}

	if wantSP.CellArea != gotSP.CellArea {
		t.Errorf("wanted CellArea %f but got CellArea %f", wantSP.CellArea, gotSP.CellArea)
	}

	if wantSP.SpIrrArea.Float64 != Utils.RoundTo(gotSP.SpIrrArea.Float64, 2) {
		t.Errorf("wanted NP Irrigated area of %f but got %f", wantSP.SpIrrArea.Float64, Utils.RoundTo(gotSP.SpIrrArea.Float64, 2))
	}

	if wantSP.SpDryArea.Float64 != Utils.RoundTo(gotSP.SpDryArea.Float64, 2) {
		t.Errorf("wanted NP Dry area of %f but got %f", wantSP.SpDryArea.Float64, Utils.RoundTo(gotSP.SpDryArea.Float64, 2))
	}

}
