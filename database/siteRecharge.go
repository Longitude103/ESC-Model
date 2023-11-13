package database

import (
	"errors"
	"fmt"
	"time"
)

type RechargeSite struct {
	Id       int     `db:"id"`
	Node     int     `db:"node"`
	Size     float64 `db:"cell_size"`
	SiteName string  `db:"site_name"`
	FileType int
	RchData  []RechargeSiteData
}

type RechargeSiteData struct {
	SiteId   int       `db:"site_id"`
	RchMonth time.Time `db:"rch_month"`
	Value    float64   `db:"value"`
}

func GetRechargeSites(v *Setup) (rechargeSites []RechargeSite, err error) {
	locationQuery := fmt.Sprintf("select recharge_sites.id id, site_name, node, st_area(mc.geom)/43560 cell_size from sp.recharge_sites inner join model_cells mc on st_contains(mc.geom, recharge_sites.geom) where mc.cell_type = %d;", v.CellType())

	if err := v.PgDb.Select(&rechargeSites, locationQuery); err != nil {
		return rechargeSites, errors.New("error getting data from recharge_sites table from DB: " + err.Error())
	}

	for i := 0; i < len(rechargeSites); i++ {
		rechargeSites[i].FileType = 117
	}

	return rechargeSites, nil
}

func GetRechargeData(v *Setup) (outputData []RechargeSiteData, err error) {
	rechargeQuery := "SELECT site_id, DATE_TRUNC('month', recharge_amount.rch_date) AS rch_month, SUM(value) AS value FROM sp.recharge_amount GROUP BY site_id, DATE_TRUNC('month',recharge_amount.rch_date);"

	var rechargeData []RechargeSiteData
	if err := v.PgDb.Select(&rechargeData, rechargeQuery); err != nil {
		return rechargeData, errors.New("error getting data from recharge_sites table from DB: " + err.Error())
	}

	// Only return recharge data between start and end dates
	for i := 0; i < len(rechargeData); i++ {
		if rechargeData[i].RchMonth.After(time.Date(v.SYear, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)) && rechargeData[i].RchMonth.Before(time.Date(v.EYear, 12, 31, 0, 0, 0, 0, time.UTC)) {
			outputData = append(outputData, rechargeData[i])
		}
	}

	return outputData, nil
}

func GetRechargeSitesWithData(v *Setup) (rechargeSites []RechargeSite, err error) {
	sites, err := GetRechargeSites(v)
	if err != nil {
		return rechargeSites, err
	}

	data, err := GetRechargeData(v)
	if err != nil {
		return rechargeSites, err
	}

	for i := 0; i < len(sites); i++ {
		for j := 0; j < len(data); j++ {
			if sites[i].Id == data[j].SiteId {
				sites[i].RchData = append(sites[i].RchData, data[j])
			}
		}
	}

	// remove recharge sites with no data
	for i := 0; i < len(sites); i++ {
		if len(sites[i].RchData) != 0 {
			rechargeSites = append(rechargeSites, sites[i])
		}
	}

	return rechargeSites, nil
}
