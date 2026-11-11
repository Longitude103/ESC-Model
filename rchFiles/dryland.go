package rchFiles

import (
	"errors"
	"fmt"
	"github.com/Longitude103/ESC-Model/database"
	"github.com/Longitude103/ESC-Model/parcels"
	"github.com/pterm/pterm"
	"time"
)

// Dryland gets a slice of dryland parcels and writes out the values to the results' database.
func Dryland(v *database.Setup, dryParcels []parcels.Parcel, cCData []database.CoeffCrop) error {
	startYear := v.SYear
	if v.SteadyState {
		if v.SYear < 1895 {
			startYear = 1895
		}
	}

	p, _ := pterm.DefaultProgressbar.WithTotal(v.EYear - startYear + 1).WithTitle("Dryland Recharge Results").WithRemoveWhenDone(true).Start()
	for y := startYear; y < v.EYear+1; y++ {
		p.UpdateTitle(fmt.Sprintf("Getting %d cells and filtering them", y))
		dryCells := database.GetDryCells(v, y) // will need to iterate through years
		annParcels, err := parcelFilterByYear(dryParcels, y)
		if err != nil {
			v.Logger.Errorf("Didn't return parcels from filter by year for year: %d", y)
			return err
		}

		p.UpdateTitle(fmt.Sprintf("Writing %d values to DB", y))
		parcelIdx := indexParcels(annParcels)
		for i := 0; i < len(dryCells); i++ {
			rows, ok := parcelIdx[parcelKey{id: int(dryCells[i].PId), nrd: dryCells[i].Nrd}]
			if !ok {
				err := errors.New("parcel not found")
				v.Logger.Errorf("Dryland RCH parcelValues error, year: %d, parcel_id: %d, nrd: %s", y,
					int(dryCells[i].PId), dryCells[i].Nrd)
				return err
			}

			row, parcelArea, _ := cellParcelRow(rows, dryCells[i].Soil, dryCells[i].CZone)
			rf, err := parcelValues(row, dryCells[i].GetLossFactor(), cCData)
			if err != nil {
				v.Logger.Errorf("Dryland RCH parcelValues error, year: %d, parcel_id: %d, nrd: %s", y,
					int(dryCells[i].PId), dryCells[i].Nrd)
				return err
			}

			for m := 0; m < 12; m++ {
				if rf[m] > 0 {
					if err := v.RchDb.Add(database.RchResult{Node: dryCells[i].Node, Size: dryCells[i].CellArea,
						Dt:       time.Date(y, time.Month(m+1), 1, 0, 0, 0, 0, time.UTC),
						FileType: 101, Result: rf[m] * dryCells[i].DryArea / parcelArea}); err != nil {
						return err
					}
				}
			}
		}

		p.Increment()
	}

	return nil
}

// parcelValues is a function that returns the monthly return flow (rf) values of a parcel row for processing.
func parcelValues(p parcels.Parcel, lossFactor float64, cCData []database.CoeffCrop) (rf [12]float64, err error) {
	etAdj, etAdjToRo, perToRch, aDp, aRo, err := database.FilterCCDryLand(cCData, p.CoeffZone, int(p.Crop1.Int64))
	if err != nil {
		return rf, err
	}

	for m := 0; m < 12; m++ {
		diffRo, diffDp := calcDiffEt(p.Et[m], etAdj, etAdjToRo)
		_, roToRch := calcRoDryland(p.Ro[m], diffRo, aRo, lossFactor, perToRch)

		rf[m] = calcDpDryland(p.Dp[m], diffDp, roToRch, aDp)
	}

	return rf, nil
}

func calcRoDryland(Ro1 float64, diffRo float64, aRo float64, lossFactor float64, perToRch float64) (runOff float64, roToRch float64) {

	totalRunOff := (Ro1 + diffRo) * aRo

	roToRch = runOffToRch(runOff, lossFactor, perToRch)
	runOff = totalRunOff * lossFactor

	return
}

func calcDpDryland(Dp1 float64, diffDp float64, roToRch float64, aDp float64) float64 {
	deepPerc := (Dp1 + diffDp) * aDp

	return deepPerc + roToRch
}
