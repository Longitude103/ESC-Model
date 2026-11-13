package rchFiles

import (
	"fmt"
	"time"

	"github.com/Longitude103/ESC-Model/database"
	"github.com/Longitude103/ESC-Model/fileio"
	"github.com/Longitude103/ESC-Model/parcels"
	"github.com/pterm/pterm"
)

func NaturalVegSS(v *database.Setup, wStations []database.WeatherStation, csResults map[string][]fileio.StationResults, cCoefficients []database.CoeffCrop) error {
	v.Logger.Info("Starting SS Natural Vegetation")

	// this needs to cycle for two years, then monthly from 1895 -> 1952
	// 2 stress periods for the first ones and then 696 months = 698 periods

	p, _ := pterm.DefaultProgressbar.WithTotal(v.EYear + 1 - v.SYear).WithTitle("Steady State Natural Vegetation Operations").WithRemoveWhenDone(true).Start()

	p.UpdateTitle("Getting 1st cell areas")
	cells1, err := database.GetSSCellAreas1(v)
	if err != nil {
		return err
	}

	p.UpdateTitle("Getting 2nd cell areas")
	cells2, err := database.GetSSCellAreas2(v)
	if err != nil {
		return err
	}

	for yr := v.SYear; yr < v.EYear+1; yr++ {
		p.UpdateTitle(fmt.Sprintf("Calculating %d SS Natural Veg Recharge", yr))
		// if the period is 0 or 1, no other parcels, just the model cells
		var cells []database.CellIntersect
		if yr < 1895 {
			cells = cells1
		} else {
			cells = cells2
		}

		for i := 0; i < len(cells); i++ {
			var cellResult []database.RchResult
			vegArea := cells[i].VegArea()

			for m := 0; m < 12; m++ {
				cellResult = append(cellResult, database.RchResult{Node: cells[i].Node, Size: cells[i].CellArea,
					Dt: time.Date(yr, time.Month(m+1), 1, 0, 0, 0, 0, time.UTC), FileType: cells[i].GetNaturalVegFileType()})
			}

			dist, err := database.Distances(cells[i], wStations)
			if err != nil {
				v.Logger.Errorf("error in distance calculation for SS cell: %v", cells[i])
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
					if data.Soil == cells[i].Soil && data.Irrigation == int(parcels.DryLand) && data.Crop == 13 {
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

			// p.UpdateTitle(fmt.Sprintf("Saving %d Natural Veg Save Results", yr))
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

	return nil
}
