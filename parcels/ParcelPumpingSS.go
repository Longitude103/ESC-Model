package parcels

import (
	"fmt"
	"sync"

	"github.com/Longitude103/ESC-Model/database"
	"github.com/Longitude103/ESC-Model/fileio"
	"github.com/Longitude103/ESC-Model/parcels/conveyLoss"
	"github.com/pterm/pterm"
)

// ParcelPumpSS is the main function for the parcels, it gets the usage, efficiencies, operates the surface water conveyance
// loss and then calls the surface water delivery. It also creates the parcels then calls the ParcelPumpDB method to set
// the parcel pumping, it then loops through the years for each parcel and sends the diversions, calls parcel NIR, sets the
// efficiency for the parcel, adds SW delivery, adds the known pumping, and then simulates pumping for all other
// parcels. Finally, it writes out the pumping per parcel and then operates the WSPP routine to finish the RO and DP.
func ParcelPumpSS(v *database.Setup, csResults map[string][]fileio.StationResults,
	wStations []database.WeatherStation, cCrops []database.CoeffCrop) (AllParcels []Parcel, err error) {

	spinner, _ := pterm.DefaultSpinner.Start("Getting Efficiencies")

	v.Logger.Info("Getting Efficiencies")
	efficiencies := database.GetSSAppEfficiency()
	spinner.Success()

	v.Logger.Info("Running Conveyance Loss")
	err = conveyLoss.Conveyance(v)
	if err != nil {
		v.Logger.Errorf("Error in Conveyance Losses %s", err)
	}

	spinner, _ = pterm.DefaultSpinner.Start("Getting Surface Water Delivery")
	// parcel delivery
	v.Logger.Info("Getting Surface Water Delivery")
	swDelivery, err := conveyLoss.GetSurfaceWaterDelivery(v)
	if err != nil {
		spinner.Fail("Error in Surface Water Delivery")
		return nil, err
	}
	spinner.Success()

	var parcels []Parcel

	// adjust start year
	startYear := 1895
	if v.SYear > 1895 {
		startYear = v.SYear
	}

	// 1. load parcels
	p, _ := pterm.DefaultProgressbar.WithTotal(v.EYear - startYear + 1).WithTitle("Parcel Operations").WithRemoveWhenDone(true).Start()
	wg := sync.WaitGroup{}

	for y := startYear; y < v.EYear+1; y++ {
		p.UpdateTitle(fmt.Sprintf("Getting %d Parels", y))
		parcels = GetParcels(v, y)

		p.UpdateTitle(fmt.Sprintf("Calculating %d Parcel NIR", y))
		for i := 0; i < len(parcels); i++ {
			wg.Add(1)
			go func(ip int) {
				defer wg.Done()
				err := (&parcels[ip]).ParcelNIR(v, y, wStations, csResults, Irrigated)
				if err != nil {
					v.Logger.Errorf("Parcel NIR Error: %s", err)
					v.Logger.Errorf("Parcel Trace: %+v", parcels[ip])
				}

				(&parcels[ip]).SetAppEfficiency(efficiencies, y)

				// add SW Delivery to the parcels
				if parcels[ip].Sw.Bool {
					(&parcels[ip]).parcelSWDelivery(swDelivery[y])
				}
			}(i) // must be a pointer to work

			wg.Wait()
		}

		v.Logger.Infof("Simulating parcel WSPP for year %d", y)
		p.UpdateTitle(fmt.Sprintf("Calculating %d Parcel WSPP", y))
		for p := 0; p < len(parcels); p++ {
			err := (&parcels[p]).WaterBalanceWSPP(v)
			if err != nil {
				v.Logger.Errorf("error in parcel WSPP: %s\n", err)
				v.Logger.Errorf("Parcel trace: %+v\n", parcels[p])
			}

			AllParcels = append(AllParcels, parcels[p])
		}
		//wg.Wait()
		p.Increment()
	}

	return AllParcels, nil
}
