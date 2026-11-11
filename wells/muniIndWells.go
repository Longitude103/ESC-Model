package wells

import (
	"time"

	"github.com/Longitude103/ESC-Model/Utils"
	"github.com/Longitude103/ESC-Model/database"
	"github.com/pterm/pterm"
)

type resultDatabase interface {
	Add(value interface{}) error
}

// MunicipalIndWells is a function that adds the municipal and industrial wells from postgresql to the results database
// and uses either assumed pumping rates or actual pumping numbers.
func MunicipalIndWells(v *database.Setup, welDB resultDatabase) error {
	spin, _ := pterm.DefaultSpinner.Start("Getting MI Wells Data and results DB")
	// go get the wells data
	wells, err := database.GetMIWells(v)
	if err != nil {
		return err
	}

	// start97 == false then use the "rate" to create the monthly pumping
	spin.UpdateText("Saving Municipal and Industrial Data")
	for yr := v.SYear; yr < v.EYear+1; yr++ {
		var wlResult []database.WelResult
		if yr < 1997 {
			for _, well := range wells {
				if !well.Start97 {
					wlResult = append(wlResult, constMIWell(well, Utils.TimeExt{Y: yr})...)
				}
			}
		} else {
			for _, well := range wells {
				if !well.Stop97 && !well.Start97 {
					wlResult = append(wlResult, constMIWell(well, Utils.TimeExt{Y: yr})...)
				}

				if well.Start97 {
					wlResult = append(wlResult, pumpMIWell(well, Utils.TimeExt{Y: yr}, v.Post97)...)
				}
			}
		}

		for i := 0; i < len(wlResult); i++ {
			if err := welDB.Add(wlResult[i]); err != nil {
				return err
			}
		}
	}

	spin.Success()
	return nil
}

func constMIWell(well database.MIWell, yr Utils.TimeExt) (wrList []database.WelResult) {
	annVolume := -1.0 * float64(well.Rate) * float64(yr.DaysInYear()) / 43560
	for i := 0; i < 12; i++ {
		dInMon := Utils.TimeExt{T: time.Date(yr.Y, time.Month(i+1), 1, 0, 0, 0, 0, time.UTC)}
		monthVol := annVolume * (float64(dInMon.DaysInMonth()) / float64(yr.DaysInYear()))
		wl := database.WelResult{Wellid: well.WellId, Node: well.Node, FileType: well.MIFileType(),
			Dt: time.Date(yr.Y, time.Month(i+1), 1, 0, 0, 0, 0, time.UTC), Result: monthVol}
		wrList = append(wrList, wl)
	}

	return wrList
}

func pumpMIWell(well database.MIWell, yr Utils.TimeExt, post97 bool) (wrList []database.WelResult) {
	if post97 {
		for _, p := range well.Pumping {
			newDate := time.Date(yr.Y, p.PumpDate.Month(), p.PumpDate.Day(), 0, 0, 0, 0, time.UTC)

			wl := database.WelResult{Wellid: well.WellId, Node: well.Node, FileType: well.MIFileType(),
				Dt: newDate, Result: p.Pump}
			wrList = append(wrList, wl)
		}
	} else {
		for _, p := range well.Pumping {
			if p.PumpDate.Year() == yr.Y {
				wl := database.WelResult{Wellid: well.WellId, Node: well.Node, FileType: well.MIFileType(),
					Dt: p.PumpDate, Result: p.Pump}
				wrList = append(wrList, wl)
			}
		}
	}

	return wrList
}
