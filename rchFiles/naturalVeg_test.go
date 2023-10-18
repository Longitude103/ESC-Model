package rchFiles_test

import (
	"fmt"
	"os"
	"testing"

	"bitbucket.org/longitude-103/wwumm-esc/database"
	"bitbucket.org/longitude-103/wwumm-esc/fileio"
	"bitbucket.org/longitude-103/wwumm-esc/rchFiles"
	"github.com/joho/godotenv"
)

func dbConnection() *database.Setup {
	var myEnv map[string]string
	myEnv, err := godotenv.Read("../.env")
	if err != nil {
		fmt.Println("Cannot load Env Variables:", err)
		os.Exit(1)
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

func Test_NaturalVeg(t *testing.T) {
	v := dbConnection()
	v.OldGrid = true

	wStations, err := database.GetWeatherStations(v.PgDb)
	if err != nil {
		t.Error("Error getting weather stations")
	}

	csResults, err := fileio.LoadTextFiles("../testData/CropSimOutput/", v.Logger)
	if err != nil {
		t.Error("Error getting CS Results")
	}

	cCoeff, err := database.GetCoeffCrops(v)
	if err != nil {
		t.Error("Error getting Crop Coefficients")
	}

	err = rchFiles.NaturalVeg(v, wStations, csResults, cCoeff)
	if err != nil {
		t.Error("Error in Natural Veg")
	}

}
