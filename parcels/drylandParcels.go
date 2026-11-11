package parcels

import (
	"fmt"
	"sync"

	"github.com/Longitude103/ESC-Model/database"
	"github.com/Longitude103/ESC-Model/fileio"
	"github.com/pterm/pterm"
)

// DryLandParcels is a function that returns all the dryland parcels for the years of the simulation and also calls the
// methods to determine parcelNIR and dryland WSPP
func DryLandParcels(v *database.Setup, csResults map[string][]fileio.StationResults,
	wStations []database.WeatherStation, cCrop []database.CoeffCrop) (dryParcels []Parcel, err error) {

	startYear := v.SYear
	if v.SteadyState {
		if v.SYear < 1895 {
			startYear = 1895
		}
	}

	p, _ := pterm.DefaultProgressbar.WithTotal(v.EYear - startYear + 1).WithTitle("Dryland Parcel Operations").WithRemoveWhenDone(true).Start()
	v.Logger.Info("Getting Dryland parcels")
	for y := startYear; y < v.EYear+1; y++ {
		p.UpdateTitle(fmt.Sprintf("Getting %d Dryland Parcels", y))
		annDryParcels := GetDryParcels(v, y)

		// method is used to set RO and DP, just poorly named.
		p.UpdateTitle(fmt.Sprintf("Calculating %d Dryland Parcels return flows", y))
		wg := sync.WaitGroup{}
		for i := 0; i < len(annDryParcels); i++ {
			wg.Add(1)
			go func(d int) {
				defer wg.Done()
				err := (&annDryParcels[d]).ParcelNIR(v, y, wStations, csResults, DryLand)
				if err != nil {
					v.Logger.Errorf("error in dry parcel NIR: %s", err)
					v.Logger.Errorf("Parcel trace: %+v", annDryParcels[d])
				}
			}(i)

			//go func(d int) {
			//	defer wg.Done()
			//	err := (&annDryParcels[d]).dryWaterBalanceWSPP(cCrop)
			//	if err != nil {
			//		v.Logger.Error("error in dry parcel WSPP ", err)
			//		v.Logger.Errorf("Parcel trace: %+v", annDryParcels[d])
			//	}
			//}(i)
		}
		wg.Wait()
		dryParcels = append(dryParcels, annDryParcels...)
		p.Increment()
	}

	v.Logger.Info("Finished Dryland parcel operations")
	return dryParcels, nil
}
