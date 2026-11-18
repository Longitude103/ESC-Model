package rchFiles

import (
	"fmt"
	"sort"
	"time"

	"github.com/Longitude103/ESC-Model/database"
	"github.com/Longitude103/ESC-Model/fileio"
	"github.com/Longitude103/ESC-Model/parcels"
)

// naturalVegPerAcre returns the natural vegetation (dryland pasture, crop 13) recharge from one acre of the cell for
// each month of yr. The recharge is linear in area, so the cell's results are this times the area of each file type.
// When matchYear is false (steady state) the first CropSim year for the station and soil is used.
func naturalVegPerAcre(v *database.Setup, cell database.CellIntersect, yr int, matchYear bool,
	wStations []database.WeatherStation, csResults map[string][]fileio.StationResults,
	cCoefficients []database.CoeffCrop) (perAcre [12]float64, err error) {

	dist, err := database.Distances(cell, wStations)
	if err != nil {
		v.Logger.Errorf("error in distance calculation for cell: %v", cell)
		return perAcre, err
	}

	etAdj, etAdjToRo, perToRch, aDp, aRo, err := database.FilterCCDryLand(cCoefficients, cell.CZone, 13)
	if err != nil {
		v.Logger.Errorf("error in getting FilterCCDryLand Function for cell: %v and crop 13", cell.CZone)
		return perAcre, err
	}

	lossFactor := cell.GetLossFactor()
	for _, st := range dist {
		var annData fileio.StationResults
		for _, data := range csResults[st.Station] {
			if (!matchYear || data.Yr == yr) && data.Soil == cell.Soil &&
				data.Irrigation == int(parcels.DryLand) && data.Crop == 13 {
				annData = data
				break
			}
		}

		if annData.Station == "" || len(annData.MonthlyData) < 12 {
			return perAcre, fmt.Errorf("no CropSim dryland pasture (crop 13) results for station %s, soil %d, year %d (cell %d)",
				st.Station, cell.Soil, yr, cell.Node)
		}

		for m := 0; m < 12; m++ {
			diffRo, diffDp := calcDiffEt(annData.MonthlyData[m].Et, etAdj, etAdjToRo)
			_, roToRch := calcRo(annData.MonthlyData[m].Ro, diffRo, st.Weight, 1, aRo, lossFactor, perToRch)
			perAcre[m] += calcDp(annData.MonthlyData[m].Dp, diffDp, roToRch, st.Weight, 1, aDp)
		}
	}

	return perAcre, nil
}

// naturalVegResults turns the per acre recharge of a cell into recharge results for each natural vegetation file type
// (NP, SP and other) by the area of natural vegetation of that type in the cell. Only results above zero are returned.
func naturalVegResults(cell database.CellIntersect, yr int, perAcre [12]float64) []database.RchResult {
	areas := cell.VegAreasByFileType()

	fileTypes := make([]int, 0, len(areas))
	for ft := range areas {
		fileTypes = append(fileTypes, ft)
	}
	sort.Ints(fileTypes)

	var results []database.RchResult
	for _, ft := range fileTypes {
		for m := 0; m < 12; m++ {
			r := perAcre[m] * areas[ft]
			if r > 0 {
				results = append(results, database.RchResult{Node: cell.Node, Size: cell.CellArea,
					Dt: time.Date(yr, time.Month(m+1), 1, 0, 0, 0, 0, time.UTC), FileType: ft, Result: r})
			}
		}
	}

	return results
}

// naturalVegCell calculates the natural vegetation recharge results of one cell for a year and adds them to the
// recharge database (debug mode only logs them).
func naturalVegCell(v *database.Setup, cell database.CellIntersect, yr int, matchYear bool,
	wStations []database.WeatherStation, csResults map[string][]fileio.StationResults,
	cCoefficients []database.CoeffCrop) error {

	if cell.VegArea() <= 0 {
		return nil
	}

	perAcre, err := naturalVegPerAcre(v, cell, yr, matchYear, wStations, csResults, cCoefficients)
	if err != nil {
		return err
	}

	for _, r := range naturalVegResults(cell, yr, perAcre) {
		if v.AppDebug {
			if r.Node == 51763 && r.Dt.Month() == time.July {
				v.Logger.Debugf("Cell: %+v, Result: %+v", cell, r)
			}
			continue
		}

		if err := v.RchDb.Add(r); err != nil {
			v.Logger.Errorf("Error Adding Result to RchDB Buffer, Result: %+v", r)
			return err
		}
	}

	return nil
}
