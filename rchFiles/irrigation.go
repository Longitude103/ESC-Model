package rchFiles

import (
	"bitbucket.org/longitude-103/wwumm-esc/database"
	"bitbucket.org/longitude-103/wwumm-esc/parcels"
	"errors"
	"fmt"
	"github.com/pterm/pterm"
	"time"
)

// IrrigationRCH is a method that creates the RCH file information in the results DB for the irrigated parcels. This uses the
// parcel information and adds the proper type id
func IrrigationRCH(v *database.Setup, AllParcels []parcels.Parcel, cCData []database.CoeffCrop) error {
	v.Logger.Info("Starting to write RCH information from Irrigated Parcels")
	p, _ := pterm.DefaultProgressbar.WithTotal(v.EYear - v.SYear + 1).WithTitle("Irrigated Recharge Results").WithRemoveWhenDone(true).Start()

	for y := v.SYear; y < v.EYear+1; y++ {
		// filter all parcels to this year only
		p.UpdateTitle(fmt.Sprintf("Filtering %d Parcels", y))
		parcelList, err := parcelFilterByYear(AllParcels, y)
		if err != nil {
			if err.Error() == "no parcels found for that year" {
				continue
			}
			v.Logger.Errorf("parcelFilterByYear error for year: %d", y)
			return err
		}

		p.UpdateTitle(fmt.Sprintf("Getting %d Irr Cells", y))
		irrCells, err := database.GetCellsIrr(v, y)
		if err != nil {
			v.Logger.Errorf("GetCellsIrr error for year: %d", y)
			return err
		}

		p.UpdateTitle(fmt.Sprintf("Saving %d Irr Cell Data", y))
		// use the RO + DP from parcel and split by acres to get recharge, will need to keep separate files for the various
		// distributions of scenarios.
		for i := 0; i < len(irrCells); i++ {
			p, err := parcelFilterById(parcelList, irrCells[i].ParcelId, irrCells[i].Nrd)
			if err != nil {
				v.Logger.Errorf("parcelFilterById error for parcel Id: %d, year: %d, and nrd: %s", irrCells[i].ParcelId, y, irrCells[i].Nrd)
				return err
			}
			fileType, err := assignRCHType(p.Nrd, p.Sw.Bool, p.Gw.Bool, p.IsPost97())
			if err != nil {
				v.Logger.Errorf("assignRCHType error where nrd is %s, SW: %t, GW: %t, post97: %t", p.Nrd, p.Sw.Bool, p.Gw.Bool, p.IsPost97())
				v.Logger.Errorf("Parcel trace: %+v", p)
				return err
			}

			_, _, perToRch, _, _, _ := database.FilterCCDryLand(cCData, p.CoeffZone, int(p.Crop1.Int64))

			cellRecharge := [12]float64{}
			for j := 0; j < 12; j++ {
				cellRecharge[j], err = cellRCH(p.Ro[j], p.Dp[j], p.Area, irrCells[i].IrrArea, irrCells[i].GetLossFactor(), perToRch)
				if err != nil {
					v.Logger.Errorf("error in cellRCH where RO: %f, DP: %f, Area: %f, IrrArea: %f", p.Ro[j], p.Dp[j], p.Area, irrCells[i].IrrArea)
					return err
				}

				if cellRecharge[j] > 0 {
					err = v.RchDb.Add(database.RchResult{Node: irrCells[i].Node, Size: irrCells[i].CellArea,
						Dt:       time.Date(y, time.Month(j+1), 1, 0, 0, 0, 0, time.UTC),
						FileType: fileType, Result: cellRecharge[j]})
					if err != nil {
						v.Logger.Errorf("Cannot Add to RchDb: %+v", database.RchResult{Node: irrCells[i].Node,
							Dt:       time.Date(y, time.Month(j+1), 1, 0, 0, 0, 0, time.UTC),
							FileType: fileType, Result: cellRecharge[j]})
						return err
					}
				}

			}
		}
		p.Increment()
	}

	v.Logger.Info("IrrigationRCH is completed.")
	return nil
}

// cellRCH returns the cell area proportion of the RCH
func cellRCH(ro float64, dp float64, parcelArea float64, parcelInCellArea float64, lossFactor float64, perToRch float64) (r float64, err error) {
	if parcelArea <= 0 {
		return 0.0, errors.New("total parcel area is zero, division by zero would occur")
	}
	roToRch := ro * lossFactor * perToRch
	r = (dp + roToRch) * parcelInCellArea / parcelArea

	return r, nil
}

// assignRCHType is a function to set the RCH Type int for the results in the database.
func assignRCHType(nrd string, sw bool, gw bool, post97 bool) (int, error) {
	switch {
	case sw && gw:
		// co-mingled
		if nrd == "sp" {
			if post97 {
				return 108, nil
			}
			return 107, nil
		}
		if post97 {
			return 106, nil
		}
		return 105, nil
	case sw && gw == false:
		// sw only
		if nrd == "sp" {
			return 104, nil
		}
		return 103, nil
	case sw == false && gw:
		// gw only
		if nrd == "sp" {
			if post97 {
				return 112, nil
			}
			return 111, nil
		}
		if post97 {
			return 110, nil
		}
		return 109, nil
	default:
		return 100, errors.New("cannot classify recharge")
	}
}
