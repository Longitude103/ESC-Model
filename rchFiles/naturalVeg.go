package rchFiles

import (
	"errors"
	"fmt"

	"github.com/Longitude103/ESC-Model/database"
	"github.com/Longitude103/ESC-Model/fileio"
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
			if err := naturalVegCell(v, cells[i], yr, true, wStations, csResults, cCoefficients); err != nil {
				return err
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

	// runOffToRch applies the loss factor itself, so it takes the total runoff (it was previously passed runOff
	// before runOff was assigned, which made the recharge from runoff always zero)
	roToRch = runOffToRch(totalRunOff, lossFactor, perToRch)
	runOff = totalRunOff * lossFactor

	return
}

func calcDp(Dp1 float64, diffDp float64, roToRch float64, weight float64, area float64, aDp float64) float64 {
	deepPerc := (Dp1 + diffDp) * weight * area / 12 * aDp

	return deepPerc + roToRch
}
