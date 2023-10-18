package rchFiles

import (
	"errors"
	"fmt"
	"time"

	"bitbucket.org/longitude-103/wwumm-esc/parcels"

	"bitbucket.org/longitude-103/wwumm-esc/database"
	"bitbucket.org/longitude-103/wwumm-esc/fileio"
	"github.com/pterm/pterm"
)

// NaturalVeg is a function that calculates the area of each cell the is natural vegetation and applies the dryland pasture
// crop type to that area. It then calculates the RO and DP for that crop at that cell location and saves it out as a
// result value in the RCH file. It does use the Adjustment Factors used in previous models.
func NaturalVeg(v *database.Setup, wStations []database.WeatherStation,
	csResults map[string][]fileio.StationResults, cCoefficients []database.CoeffCrop) error {
	v.Logger.Info("Starting Natural Vegetation Ops.")

	p, _ := pterm.DefaultProgressbar.WithTotal(v.EYear - v.SYear + 1).WithTitle("Natural Vegetation Operations").WithRemoveWhenDone(true).Start()
	for yr := v.SYear; yr < v.EYear+1; yr++ {
		p.UpdateTitle(fmt.Sprintf("Getting %d cell areas", yr))
		cells, err := database.GetCellAreas(v, yr)
		if err != nil {
			e := fmt.Sprintf("error in getting cell areas for yr %d, error is: %s", yr, err)
			return errors.New(e)
		}

		p.UpdateTitle(fmt.Sprintf("Calculating %d Natural Veg Recharge", yr))
		for i := 0; i < len(cells); i++ {
			var cellResult []database.RchResult
			vegArea := cells[i].VegArea()

			for m := 0; m < 12; m++ {
				cellResult = append(cellResult, database.RchResult{Node: cells[i].Node, Size: cells[i].CellArea,
					Dt: time.Date(yr, time.Month(m+1), 1, 0, 0, 0, 0, time.UTC), FileType: 102})
			}

			dist, err := database.Distances(cells[i], wStations)
			if err != nil {
				v.Logger.Errorf("error in distance calculation for cell: %v", cells[i])
				return err
			}

			etAdj, etAdjToRo, perToRch, aDp, aRo, err := database.FilterCCDryLand(cCoefficients, cells[i].CZone, 13)
			if err != nil {
				v.Logger.Errorf("error in getting FilterCCDryLand Function for cell: %v and crop 13", cells[i].CZone)
				return err
			}

			for _, st := range dist {
				var annData fileio.StationResults
				for _, data := range csResults[st.Station] {
					if data.Yr == yr && data.Soil == cells[i].Soil &&
						data.Irrigation == int(parcels.DryLand) && data.Crop == 13 {
						annData = data
						break
					}
				}

				if annData.Station == "" {
					v.Logger.Debugf("AnnData: %+v", annData)
					v.Logger.Debugf("cell[i]: %+v", cells[i])

				}

				for m := 0; m < 12; m++ {
					diffRo, diffDp := calcDiffEt(annData.MonthlyData[m].Et, etAdj, etAdjToRo)
					_, roToRch := calcRo(annData.MonthlyData[m].Ro, diffRo, st.Weight, vegArea, aRo, cells[i].GetLossFactor(), perToRch)
					deepPerc := calcDp(annData.MonthlyData[m].Dp, diffDp, roToRch, st.Weight, vegArea, aDp)

					if v.AppDebug {
						if m == 6 && cellResult[m].Node == 90833 {
							v.Logger.Debugf("st: %+v, VegArea: %f, CellLossFactor: %f", st, vegArea, cells[i].GetLossFactor())
							v.Logger.Debugf("MonthET: %f, MonthRO: %f, MonthDP: %f", annData.MonthlyData[m].Et, annData.MonthlyData[m].Ro, annData.MonthlyData[m].Dp)
							v.Logger.Debugf("diffRo: %f, diffDp: %f", diffRo, diffDp)
							v.Logger.Debugf("roToRch: %f", roToRch)
							v.Logger.Debugf("deepPerc: %f", deepPerc)
						}
					}

					cellResult[m].Result += deepPerc
				}
			}

			p.UpdateTitle(fmt.Sprintf("Saving %d Natural Veg Save Results", yr))
			for m := 0; m < 12; m++ {
				if cellResult[m].Result > 0 {
					if v.AppDebug {
						if m == 6 && cellResult[m].Node == 51763 {
							v.Logger.Debugf("Cells Value: %+v", cells[i])
							v.Logger.Debugf("Distances: %+v", dist)
							v.Logger.Debugf("etAdj: %f, etAdjToRo: %f, perToRch: %f, aDp: %f, aRo: %f", etAdj, etAdjToRo, perToRch, aDp, aRo)
							v.Logger.Debugf("Result: %+v", cellResult)
						}
						// v.Logger.Debugf("Result: %+v", cellResult)
					} else {
						if err := v.RchDb.Add(cellResult[m]); err != nil {
							v.Logger.Errorf("Error Adding Result to RchDB Buffer, Result: %+v", cellResult)
							return err
						}
					}

				}
			}
		}
		p.Increment()
	}

	v.Logger.Info("finished natural vegetation function.")
	return nil
}

func calcDiffEt(Et float64, etAdj float64, etAdjToRo float64) (diffEtToRo float64, diffEtToDp float64) {
	if etAdj == 1 {
		return 0, 0
	}

	diffEt := Et * (1 - etAdj)       // ET that was removed, if any
	diffEtToRo = diffEt * etAdjToRo  // the difference to RO
	diffEtToDp = diffEt - diffEtToRo // Remaining to DP

	return
}

func runOffToRch(Ro float64, lossFactor float64, perToRch float64) float64 {
	lossRo := Ro * lossFactor
	return lossRo * perToRch // amount of runoff that infiltrates on the way back to stream
}

func calcRo(Ro1 float64, diffRo float64, weight float64, area float64, aRo float64, lossFactor float64,
	perToRch float64) (runOff float64, roToRch float64) {

	totalRunOff := (Ro1 + diffRo) * weight * area / 12 * aRo

	roToRch = runOffToRch(runOff, lossFactor, perToRch)
	runOff = totalRunOff * lossFactor

	return
}

func calcDp(Dp1 float64, diffDp float64, roToRch float64, weight float64, area float64, aDp float64) float64 {
	deepPerc := (Dp1 + diffDp) * weight * area / 12 * aDp

	return deepPerc + roToRch
}
