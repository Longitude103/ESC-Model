package fileio

import (
	"bitbucket.org/longitude-103/wwumm-esc/logging"
	"bufio"
	"fmt"
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
	fls, err := os.ReadDir(filePath)
	// TODO: Ensure that there are Weather files here, otherwise throw an error
	if err != nil {
		logger.Errorf("Error loading text files %s", err)
		return nil, err
	}

	dataMap := make(map[string][]StationResults)
	for _, v := range fls {
		wStationId := v.Name()
		// skip folders, hidden files (.DS_Store etc.) and anything too short to hold a station id
		if v.IsDir() || strings.HasPrefix(wStationId, ".") || len(wStationId) < 4 {
			logger.Infof("Skipping non-station entry: %s", wStationId)
			continue
		}

		logger.Infof("Reading station: %s", wStationId[:4])
		path := filepath.Join(filePath, v.Name())
		data, err := getFileData(path, logger)
		if err != nil {
			logger.Errorf("Error in processing data for %s, %s", wStationId, err)
			return nil, err
		}

		dataMap[wStationId[:4]] = data
	}

	return dataMap, nil
}

// getFileData is a function that breaks down the station data and puts it into a struct to work with.
func getFileData(filePath string, logger *logging.TheLogger) ([]StationResults, error) {
	file, err := os.Open(filePath)
	if err != nil {
		logger.Errorf("Error in getting %s file data, err: %s", filePath, err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()

	var stationData []StationResults
	lineNo := 1
	for scanner.Scan() {
		lineNo++
		elements := strings.Fields(scanner.Text())
		if len(elements) == 0 {
			continue
		}

		if len(elements) < 78 {
			return nil, fmt.Errorf("%s line %d: expected 78 fields, found %d", filePath, lineNo, len(elements))
		}

		station := StationResults{Station: elements[0]}
		ints := []*int{&station.Yr, &station.Soil, &station.Crop, &station.Tillage, &station.Irrigation}
		for i, ptr := range ints {
			if *ptr, err = strconv.Atoi(elements[i+1]); err != nil {
				return nil, fmt.Errorf("%s line %d: %w", filePath, lineNo, err)
			}
		}

		for i := 6; i < 78; i = i + 6 {
			var vals [6]float64
			for j := 0; j < 6; j++ {
				if vals[j], err = strconv.ParseFloat(elements[i+j], 64); err != nil {
					return nil, fmt.Errorf("%s line %d: %w", filePath, lineNo, err)
				}
			}

			station.MonthlyData = append(station.MonthlyData, MonthlyValues{Et: vals[0], Eff_precip: vals[1],
				Nir: vals[2], Dp: vals[3], Ro: vals[4], Precip: vals[5]})
		}

		stationData = append(stationData, station)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return stationData, nil
}
