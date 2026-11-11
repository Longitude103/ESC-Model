package wells

import (
	"github.com/Longitude103/ESC-Model/database"
	"github.com/pterm/pterm"
)

// SteadyStateWells is a function to get the steady state wells from the DB and write them into the results db for the
// model year(s) in the proper nodes and with the proper amounts.
func SteadyStateWells(v *database.Setup, welDB resultDatabase) error {
	spin, _ := pterm.DefaultSpinner.Start("Getting Steady State Wells and Pumping")
	// get wells from DB with their node number
	v.Logger.Info("Starting Steady State Wells Addition")
	ssWells, err := database.GetSSWells(v)
	if err != nil {
		return err
	}

	p, _ := pterm.DefaultProgressbar.WithTotal(v.EYear - v.SYear + 1).WithTitle("Steady State Save Results").WithRemoveWhenDone(true).Start()
	for yr := v.SYear; yr < v.EYear+1; yr++ {
		p.Increment()
		// write them to the results DB
		for i := 0; i < len(ssWells); i++ {
			if err := welDB.Add(database.WelAnnualResult{Wellid: ssWells[i].WellName, Node: ssWells[i].Node, Yr: yr,
				FileType: 209, Result: ssWells[i].MVolume}); err != nil {
				return err
			}
		}
	}

	spin.Success()
	v.Logger.Info("Added all the steady state wells to results DB")
	return nil
}
