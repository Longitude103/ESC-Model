package conveyLoss

import (
	"github.com/Longitude103/ESC-Model/database"
	"time"
)

// GetSurfaceWaterDelivery function returns a map with a key of year and a value of slice of Diversion that is a monthly amount of surface water delivered to
// an acre of land. The units of the Diversion are in acre-feet per acre for use in subsequent processes.
func GetSurfaceWaterDelivery(v *database.Setup) (map[int]Diversions, error) {
	var db *database.SWDelDB
	var err error
	if !v.AppDebug {
		db, err = database.SWDeliveryDB(v.SlDb)
		if err != nil {
			return nil, err
		}
	}

	diversions, _, err := getDiversions(v)
	if err != nil {
		v.Logger.Errorf("Error in getDiversions: %s", err)
		return nil, err
	}

	geringDivs := getGeringDivs(diversions)

	canals, err := getCanals(v)
	if err != nil {
		v.Logger.Errorf("Error in getCanals: %s", err)
		return nil, err
	}

	for i := 0; i < len(diversions); i++ {
		c := filterCnl(canals, diversions[i].CanalId, diversions[i].DivDate.Time.Year())
		//fmt.Printf("Filtered canal is: %+v\n", c)
		if c.Area.Valid {
			if diversions[i].CanalId == 13 && diversions[i].DivDate.Time.Year() > 1984 {
				// fix diversions for Mitchell Canal after 1985
				diversions[i].DivAmount.Float64 = fixMitchell(diversions[i], geringDivs, c.Eff, c.Area.Float64)
			} else {
				// apply efficiency and convert to AF
				(&diversions[i]).applyEffAcres(c.Eff, c.Area.Float64)
			}

		} else {
			(&diversions[i]).DivAmount.Float64 = 0
		}
	}

	if !v.AppDebug {
		for i := 0; i < len(diversions); i++ {
			if err := db.Add(database.SWDelResult{CanalId: diversions[i].CanalId, Dt: diversions[i].DivDate.Time,
				DelAmount: diversions[i].DivAmount.Float64}); err != nil {
				return nil, err
			}
		}

		_ = db.Flush()
		_ = db.Close()
	}

	mapDivs := make(map[int]Diversions)

	for y := v.SYear; y < v.EYear+1; y++ {
		mapDivs[y] = diversions.FilterSWDeliveryByYear(y)
	}

	return mapDivs, nil
}

func getGeringDivs(diversions Diversions) map[int]Diversions {
	var geringDiv = make(map[int]Diversions)

	for _, diversion := range diversions {
		if diversion.CanalId == 32 {
			if diversion.DivDate.Time.After(time.Date(1984, time.Month(12), 1, 0, 0, 0, 0, time.UTC)) {
				geringDiv[diversion.DivDate.Time.Year()] = append(geringDiv[diversion.DivDate.Time.Year()], diversion)
			}
		}
	}

	return geringDiv
}

func fixMitchell(diversion Diversion, geringDivs map[int]Diversions, eff float64, acres float64) float64 {
	divs := geringDivs[diversion.DivDate.Time.Year()]
	adjDiv := 0.0

	for _, div := range divs {
		if diversion.DivDate.Time.Equal(div.DivDate.Time) {
			adjDiv = (diversion.DivAmount.Float64*eff - div.DivAmount.Float64) / acres
		}
	}

	return adjDiv
}

// filterCnl filters the list of canals to a specific one.
func filterCnl(canals []Canal, canal int, yr int) Canal {
	for _, v := range canals {
		if v.Id == canal && v.Yr == yr {
			return v
		}
	}

	return Canal{}
}
