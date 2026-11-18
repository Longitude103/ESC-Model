package rchFiles

import (
	"fmt"

	"github.com/Longitude103/ESC-Model/database"
	"github.com/Longitude103/ESC-Model/fileio"
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
			if err := naturalVegCell(v, cells[i], yr, false, wStations, csResults, cCoefficients); err != nil {
				return err
			}
		}
		p.Increment()
	}

	return nil
}
