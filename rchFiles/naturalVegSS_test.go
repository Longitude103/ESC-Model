package rchFiles_test

import (
	"testing"

	"bitbucket.org/longitude-103/wwumm-esc/database"
	"bitbucket.org/longitude-103/wwumm-esc/fileio"
	"bitbucket.org/longitude-103/wwumm-esc/rchFiles"
)

func TestNaturalVegSS(t *testing.T) {
	v := dbConnection()
	v.SteadyState = true
	wStations, err := database.GetWeatherStations(v.PgDb)
	if err != nil {
		t.Error("Error getting weather stations")
	}

	csResults, err := fileio.LoadTextFiles("../testData/CropSimOutput/", v.Logger)
	if err != nil {
		t.Error("Error getting CS Results")
	}

	avgCSResults, err := fileio.AverageStationResults(csResults, 1953, 2020)
	if err != nil {
		t.Error("Error in Averaging CS Results")
	}

	cCoeff, err := database.GetCoeffCrops(v)
	if err != nil {
		t.Error("Error getting Crop Coefficients")
	}

	err = rchFiles.NaturalVegSS(v, wStations, avgCSResults, cCoeff)
	if err != nil {
		t.Error("Error in Natural Veg SS Function")
	}

}
