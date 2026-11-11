package fileio

import (
	"bufio"
	"github.com/Longitude103/ESC-Model/logging"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type StationResults struct {
	Station     string
	Soil        int
	MonthlyData []MonthlyValues
	Yr          int
	Crop        int
	Tillage     int
	Irrigation  int
}

type MonthlyValues struct {
	Et         float64
	Eff_precip float64
	Nir        float64
	Dp         float64
	Ro         float64
	Precip     float64
}

// LoadTextFiles loads CropSim text files from a location and returns a map of the results to use in further processing.
// This should include all the files and results for each station and each year. The map key is the station id that is
// also stored in the StationResults struct as Station
func LoadTextFiles(filePath string, logger *logging.TheLogger) (map[string][]StationResults, error) {
	logger.Infof("File Path: %s", filePath)
	//fmt.Println("File Path:", filePath)
	fls, err := os.ReadDir(filePath)
	// TODO: Ensure that there are Weather files here, otherwise throw an error
	if err != nil {
		logger.Errorf("Error loading text files %s", err)
		return nil, err
	}

	dataMap := make(map[string][]StationResults)
	for _, v := range fls {
		wStationId := v.Name()
		logger.Infof("Reading station: %s", wStationId[:4])
		path := filepath.Join(filePath, v.Name())
		dataMap[wStationId[:4]], err = getFileData(path, logger)
	}
	if err != nil {
		logger.Errorf("Error in processing data, %s", err)
		return nil, err
	}

	return dataMap, nil
}

// getFileData is a function that breaks down the station data and puts it into a struct to work with.
func getFileData(filePath string, logger *logging.TheLogger) ([]StationResults, error) {
	file, err := os.Open(filePath)
	if err != nil {
		logger.Errorf("Error in getting %s file data, err: %s", file.Name(), err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	//fmt.Println("Line 1:", scanner.Text())
	//fmt.Println("Line 1 Split", strings.Fields(scanner.Text()))
	//fmt.Println("Line 1 length", len(strings.Fields(scanner.Text()))) // 78

	var stationData []StationResults
	for scanner.Scan() {
		elements := strings.Fields(scanner.Text())
		station := StationResults{Station: elements[0]}
		station.Yr, err = strconv.Atoi(elements[1])
		station.Soil, err = strconv.Atoi(elements[2])
		station.Crop, err = strconv.Atoi(elements[3])
		station.Tillage, err = strconv.Atoi(elements[4])
		station.Irrigation, err = strconv.Atoi(elements[5])

		for i := 6; i < 78; i = i + 6 {
			var mvals MonthlyValues
			mvals.Et, err = strconv.ParseFloat(elements[i], 64)
			mvals.Eff_precip, err = strconv.ParseFloat(elements[i+1], 64)
			mvals.Nir, err = strconv.ParseFloat(elements[i+2], 64)
			mvals.Dp, err = strconv.ParseFloat(elements[i+3], 64)
			mvals.Ro, err = strconv.ParseFloat(elements[i+4], 64)
			mvals.Precip, err = strconv.ParseFloat(elements[i+5], 64)

			station.MonthlyData = append(station.MonthlyData, mvals)
		}

		stationData = append(stationData, station)
	}
	if err != nil {
		return nil, err
	}

	return stationData, nil
}
