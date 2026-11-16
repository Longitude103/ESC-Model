package conveyLoss

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Longitude103/ESC-Model/Utils"
	"github.com/Longitude103/ESC-Model/database"
	"github.com/pterm/pterm"
)

// CanalRecharge is a function to get the recharge amounts from the diversions and apportion them to the Canal sections
// and recharge them into the groundwater model.
func CanalRecharge(v *database.Setup) (err error) {
	v.Logger.Info("Started Excess Flow Procedure")
	spinner, _ := pterm.DefaultSpinner.Start("Started Excess Flow Procedure")
	_, diversions, err := getDiversions(v)
	if err != nil {
		spinner.Fail("Get Diversions Failed")
		return err
	}

	if len(diversions) == 0 { // no recharge this year, exit function
		spinner.Success("Excess Flow Procedure Complete")
		return nil
	}

	canalCells, err := getCanalCells(v)
	if err != nil {
		spinner.Fail("Get Canals Failed")
		return err
	}

	efDivs := getLossPercent(diversions, v)

	spinner.Success()
	checkDivTotal := 0.0
	p, _ := pterm.DefaultProgressbar.WithTotal(len(canalCells)).WithTitle("Process Excess Flow Canal Cells").WithRemoveWhenDone(true).Start()
	for _, cell := range canalCells {
		p.Increment()
		if cell.CanalType != "Canal" {
			continue
		}

		factor := getFactor(cell)                                         // gets the loss percent factor for the cell.
		efCanalDiversions := efDivs.FindDiversionsByCanalId(cell.CLinkId) // filter diversions to the canal

		if v.AppDebug && cell.Node == 12980 {
			v.Logger.Debugf("Cell Properties: %s", cell.sprint())
			v.Logger.Debugf("cellIdDiv: %d", cell.CLinkId)
			v.Logger.Debugf("CellDiversions: %+v", efCanalDiversions)
		}

		structureLoss := 0.0
		for _, div := range efCanalDiversions {
			structureLoss = div.LossPercent.Float64 * div.DivAmount.Float64

			if v.AppDebug {
				if cell.Node == 12987 || cell.Node == 13006 || cell.Node == 13043 {
					v.Logger.Debugf("Div: %+v\n", div)
					v.Logger.Debugf("Structure Loss Percent: %f, StructureLoss: %f, Factor: %f\n", div.LossPercent.Float64, structureLoss, factor)
					v.Logger.Debugf("Cell Data: %+v\n", cell)
					d := database.RchResult{Node: cell.Node, Size: cell.CellArea, Dt: div.DivDate.Time,
						FileType: 124, Result: structureLoss * factor * 1.9835}
					v.Logger.Debugf("Cell Result: %+v\n", d)
				}

				if cell.CanalId == 11 && div.DivDate.Time.Equal(time.Date(2011, 4, 1, 0, 0, 0, 0, time.UTC)) {
					checkDivTotal += structureLoss * factor * 1.9835
				}
			} else {
				if structureLoss > 0 {
					ft := 124
					if div.CanalId == 402 {
						ft = 125
					}

					err := v.RchDb.Add(database.RchResult{Node: cell.Node, Size: cell.CellArea, Dt: div.DivDate.Time,
						FileType: ft, Result: structureLoss * factor * 1.9835})
					if err != nil {
						return err
					}
				}
			}
		}
	}

	if err := MakeRechargeData(v); err != nil {
		return err
	}

	if v.AppDebug {
		v.Logger.Debugf("Check Total is %f", checkDivTotal)
	}
	pterm.Success.Println("Excess Flow Calculations Complete")
	v.Logger.Info("Excess Flow Completed Successfully")
	return nil
}

type efDiversion struct {
	Diversion
	LossPercent sql.NullFloat64
}

type efDiversions []efDiversion

func (efDivs *efDiversions) FindDiversionsByCanalId(canalId int) (canalDiversions efDiversions) {
	// filterCanal filters the canal diversions to a specific canal and returns a slice of Diversion
	for _, v := range *efDivs {
		if v.CanalId == canalId {
			canalDiversions = append(canalDiversions, v)
		}
	}

	return
}

func getLossPercent(divs []Diversion, v *database.Setup) (efDiv efDiversions) {
	formattedQueries := Utils.SplitQueries(divQueries)

	// remove the excess flows
	efQuery := fmt.Sprintf(formattedQueries[3], v.SYear, v.EYear, v.SYear, v.EYear)
	var efPeriods []efPeriod
	if err := v.PgDb.Select(&efPeriods, efQuery); err != nil {
		v.Logger.Errorf("Error in getting Excess Flow Periods: %s", err)
	}

	return matchEFPeriods(divs, efPeriods)
}

// matchEFPeriods pairs each monthly excess flow diversion with the excess flow period that covers its month. The
// monthly diversion already holds every excess flow day in that month, so it is matched to one period only (the first
// that covers it); matching it to every period in the month counted the same water more than once.
func matchEFPeriods(divs []Diversion, efPeriods []efPeriod) (efDiv efDiversions) {
	for _, d := range divs {
		for _, e := range efPeriods {
			if d.CanalId != e.CanalId {
				continue
			}

			y, months := e.GetYearAndMonths()
			m := int(d.DivDate.Time.Month())
			if y == d.DivDate.Time.Year() && m >= months[0] && m <= months[len(months)-1] {
				efDiv = append(efDiv, efDiversion{
					Diversion: Diversion{
						CanalId:   d.CanalId,
						DivDate:   sql.NullTime{Time: d.DivDate.Time, Valid: d.DivDate.Valid},
						DivAmount: sql.NullFloat64{Float64: d.DivAmount.Float64, Valid: d.DivAmount.Valid},
					},
					LossPercent: e.LossPercet,
				})
				break
			}
		}
	}

	return
}
