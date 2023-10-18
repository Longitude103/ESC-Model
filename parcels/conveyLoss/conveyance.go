package conveyLoss

import (
	"time"

	"bitbucket.org/longitude-103/wwumm-esc/database"
	"github.com/pterm/pterm"
)

// Conveyance function finds the diversions and calculates the conveyance loss for all cells where there is a canal. This
// outputs to the results table in sqlite. Might update to return delivery by canal.
func Conveyance(v *database.Setup) (err error) {
	spRates := map[string]float64{"interstate": 0.4869, "highline": 0.2617, "lowline": 0.2513}

	spinner, _ := pterm.DefaultSpinner.Start("Getting Canal Cells and Diversions")
	canalCells, err := getCanalCells(v)
	if err != nil {
		spinner.Fail("Get Canals Failed")
		return err
	}

	diversions, _, err := getDiversions(v)
	if err != nil {
		spinner.Fail("Get Diversions Failed")
		return err
	}
	spinner.Success()
	checkDivTotal := 0.0

	p, _ := pterm.DefaultProgressbar.WithTotal(len(canalCells)).WithTitle("Process Canal Cells").WithRemoveWhenDone(true).Start()
	// loop over cells
	for _, cell := range canalCells {
		p.Increment()
		strLossPercent := 0.0
		cellIdDiv := 0

		// determine efficiency and get total loss factor
		if cell.CanalType == "Lateral" || cell.CanalType == "Spill" {
			strLossPercent = (1 - cell.CanalEff.Float64) * 1 / 6
			cell.CanalId = cell.CLinkId
			cellIdDiv = cell.CLinkId
		} else {
			cellIdDiv = cell.CanalId
			if cell.LatCount.Int64 == 0 {
				strLossPercent = 1 - cell.CanalEff.Float64
			} else {
				strLossPercent = (1 - cell.CanalEff.Float64) * 5 / 6
			}
		}

		factor := getFactor(cell)

		// special cases for cells with Minatare, Mitchell Gering, and Highline and Lowline
		switch cellIdDiv {
		case 29: // Use the same diversions for north and south Minatare cells
			cellIdDiv = 30
		case 13: //  Mitchell Gering use Mitchell and Gering diversions for Gering Canal
			cellIdDiv = 13
		case 21, 25: // Highline and lowline canal use interstate canal diversions split by percentage of acerage
			cellIdDiv = 2
		}

		// filter diversions to the canal
		canalDiversions := diversions.FindDiversionsByCanalId(cellIdDiv)

		if v.AppDebug && cell.Node == 15661 {
			v.Logger.Debugf("Cell Properties: %s", cell.sprint())
			v.Logger.Debugf("cellIdDiv: %d", cellIdDiv)
			v.Logger.Debugf("CellDiversions: %+v", canalDiversions)
		}

		structureLoss := 0.0
		for _, div := range canalDiversions {
			// Skip divs that are pre-1895 for Steady State
			if v.SteadyState {
				if div.DivDate.Time.Before(time.Date(1895, 1, 1, 0, 0, 0, 0, time.UTC)) {
					continue
				}
			}

			if cell.CanalId == 2 || cell.CanalId == 25 || cell.CanalId == 21 {
				switch div.DivDate.Time.Month() {
				case 1, 2, 3, 10, 11, 12:
					if cell.CanalId == 2 {
						structureLoss = strLossPercent * div.DivAmount.Float64
					} else {
						structureLoss = 0.0
					}
				default:
					switch cell.CanalId {
					case 2:
						structureLoss = strLossPercent * div.DivAmount.Float64 * spRates["interstate"]
					case 21:
						structureLoss = strLossPercent * div.DivAmount.Float64 * spRates["lowline"]
					case 25:
						structureLoss = strLossPercent * div.DivAmount.Float64 * spRates["highline"]
					}
				}
			} else {
				structureLoss = strLossPercent * div.DivAmount.Float64
			}

			ft := 113               // np by default
			if cell.CanalId == 54 { // western canal, only sp canal
				ft = 114
			}

			if v.AppDebug {
				if cell.Node == 17467 || cell.Node == 17468 || cell.Node == 7601 {
					v.Logger.Debugf("Div: %+v\n", div)
					v.Logger.Debugf("Structure Loss Percent: %f, StructureLoss: %f, Factor: %f\n", strLossPercent, structureLoss, factor)
					v.Logger.Debugf("Cell Data: %+v\n", cell)
					d := database.RchResult{Node: cell.Node, Size: cell.CellArea, Dt: div.DivDate.Time,
						FileType: ft, Result: structureLoss * factor * 1.9835}
					v.Logger.Debugf("Cell Result: %+v\n", d)
				}

				if cell.CanalId == 2 && div.DivDate.Time.Equal(time.Date(1953, 4, 1, 0, 0, 0, 0, time.UTC)) {
					checkDivTotal += structureLoss * factor * 1.9835
				}
			} else {
				if structureLoss > 0 {
					err := v.RchDb.Add(database.RchResult{Node: cell.Node, Size: cell.CellArea, Dt: div.DivDate.Time,
						FileType: ft, Result: structureLoss * factor * 1.9835})
					if err != nil {
						return err
					}
				}
			}
		}
	}

	if v.AppDebug {
		v.Logger.Debugf("Check Total is %f", checkDivTotal)
	}
	pterm.Success.Println("Canal Loss Calculations")
	v.Logger.Info("Canal Loss Completed Successfully")
	return nil
}

// getFactor is a function that returns the loss factor of the cell based on the "CFlag" of that cell. I can also give a default value
// if there is not a flag set or one of the flags is missing.
func getFactor(cell CanalCell) (factor float64) {
	// switch cell.CFlag {
	// case 1: // DNR Factor
	// 	if !cell.DnrFact.Valid {
	// 		factor = defaultFactor(cell)
	// 	} else {
	// 		factor = cell.DnrFact.Float64
	// 	}
	// case 4: // SatThick Factor
	// 	if !cell.SatFact.Valid {
	// 		factor = defaultFactor(cell)
	// 	} else {
	// 		factor = cell.SatFact.Float64
	// 	}
	// case 2: // USGS Factor
	// 	if !cell.UsgsFact.Valid {
	// 		factor = defaultFactor(cell)
	// 	} else {
	// 		factor = cell.UsgsFact.Float64
	// 	}
	// default: // default
	// 	factor = defaultFactor(cell)
	// }

	return defaultFactor(cell)
}

// defaultFactor is a function that returns the default factor of the canal which is the portion of the canal length within that
// cell and or the portion of lateral length within the cell divided by the total lateral length.
func defaultFactor(cell CanalCell) (factor float64) {
	if cell.CanalType == "Lateral" || cell.CanalType == "Spill" {
		factor = cell.StLength / cell.TotalLatLn.Float64
	} else {
		factor = cell.StLength / cell.TotalCanLn
	}

	return factor
}
