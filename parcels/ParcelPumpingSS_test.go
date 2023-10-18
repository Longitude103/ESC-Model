package parcels_test

import (
	"bitbucket.org/longitude-103/wwumm-esc/database"
	"bitbucket.org/longitude-103/wwumm-esc/fileio"
	"bitbucket.org/longitude-103/wwumm-esc/parcels"
	"testing"
)

func TestParcelPumpSS(t *testing.T) {
	v := dbConnection()
	v.SYear = 1895
	v.EYear = 1895
	v.SteadyState = true

	csResults, err := fileio.LoadTextFiles("../testData/CropSimOutput/", v.Logger)
	if err != nil {
		t.Error("Error getting CS Results")
	}

	avgCSResults, err := fileio.AverageStationResults(csResults, 1953, 2020)
	if err != nil {
		t.Error("Error in Averaging CS Results")
	}

	v.Logger.Info("Getting Weather Stations")
	wStations, err := database.GetWeatherStations(v.PgDb)
	if err != nil {
		t.Error("Error getting weather stations")
	}

	cCoefficients, err := database.GetCoeffCrops(v)
	if err != nil {
		t.Error("Error getting coefficient of crops")
	}

	ap, err := parcels.ParcelPumpSS(v, avgCSResults, wStations, cCoefficients)
	if err != nil {
		t.Errorf("Error in parcel pumping SS process: %s", err)
	}

	for _, a := range ap {
		if a.Sw.Bool == false {
			t.Errorf("parcel does not have SW: %+v\n", a)
		}

		if a.Yr > 1952 || a.Yr < 1895 {
			t.Error("parcels in the wrong year range for Steady State")
		}
	}
}
