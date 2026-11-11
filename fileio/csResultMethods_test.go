package fileio_test

import (
	"fmt"
	"testing"

	"github.com/Longitude103/ESC-Model/database"
	"github.com/Longitude103/ESC-Model/fileio"
	"github.com/joho/godotenv"
)

func dbConnection(t testing.TB) *database.Setup {
	t.Helper()
	var myEnv map[string]string
	myEnv, err := godotenv.Read("../.env")
	if err != nil {
		t.Skipf("skipping database test: cannot load .env: %v", err)
	}

	var v *database.Setup
	v, err = database.NewSetup(myEnv, database.WithNoSQLite(), database.WithDebug(), database.WithOldGrid())
	if err != nil {
		fmt.Printf("Error in NewSetup: %s", err)
	}

	if err = v.SetYears(1953, 1954); err != nil {
		fmt.Println("error setting years")
	}

	return v
}

func TestAverageStationsResults(t *testing.T) {
	v := dbConnection(t)
	v.SteadyState = true

	csResults, err := fileio.LoadTextFiles("../testData/CropSimOutput/", v.Logger)
	if err != nil {
		t.Error("Error getting CS Results")
	}

	avgCSResults, err := fileio.AverageStationResults(csResults, 1953, 2020)
	if err != nil {
		t.Error("Error in Averaging CS Results")
	}

	// fmt.Printf("CS Results: %+v", avgCSResults)
	fmt.Printf("Number of Stations in avgCSResults %d\n", len(avgCSResults))
	Oshk := avgCSResults["OSHK"]

	var foundOshk bool
	for _, o := range Oshk {
		// fmt.Printf("Year: %d; Soil: %d, Crop: %d, Tillage: %d, Irr: %d\n", o.Yr, o.Soil, o.Crop, o.Tillage, o.Irrigation)
		if o.Yr == 1952 && o.Soil == 412 && o.Crop == 13 && o.Tillage == 1 && o.Irrigation == 1 {
			foundOshk = true
		}
	}

	if !foundOshk {
		t.Error("Didn't find the native pasture for soil 412 in OSHK")
	}
}
