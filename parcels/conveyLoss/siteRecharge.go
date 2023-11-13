package conveyLoss

import (
	"bitbucket.org/longitude-103/wwumm-esc/database"
)

// get recharge data from postgres database
// aggregate daily data to monthly data
// run spatial query to find out what model grid the recharge site is located in
// assign new recharge data by model grid into Results table of results database

func MakeRechargeData(v *database.Setup) (err error) {
	rechargeSites, err := database.GetRechargeSitesWithData(v)
	if err != nil {
		v.Logger.Errorf("Error in getting recharge data from DB: %s", err)
		return err
	}

	rchDB, err := database.ResultsRchDB(v.SlDb)
	if err != nil {
		v.Logger.Errorf("Error in getting Results RCH DB: %s", err)
		return err
	}

	v.Logger.Info("Saving the SP Recharge Site information to results database")
	for i := 0; i < len(rechargeSites); i++ {
		for j := 0; j < len(rechargeSites[i].RchData); j++ {
			if err := rchDB.Add(database.RchResult{Node: rechargeSites[i].Node, Size: rechargeSites[i].Size, Dt: rechargeSites[i].RchData[j].RchMonth,
				FileType: rechargeSites[i].FileType, Result: rechargeSites[i].RchData[j].Value}); err != nil {
				v.Logger.Errorf("Error in adding recharge data to results database: %s", err)
				return err
			}
		}
	}

	_ = rchDB.Flush()
	v.Logger.Info("SP Recharge Site function completed.")
	return nil
}
