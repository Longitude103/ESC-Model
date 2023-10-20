package database

import (
	"errors"
	"fmt"
	"time"
)

type RechargeSite struct {
	id       int     `db:"id"`
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
	locationQuery := fmt.Sprintf("select recharge_sites.id, site_name, node, st_area(mc.geom)/43560 cell_size from sp.recharge_sites inner join model_cells mc on st_contains(mc.geom, recharge_sites.geom) where mc.cell_type = %d;", v.CellType())

	if err := v.PgDb.Select(&rechargeSites, locationQuery); err != nil {
		return rechargeSites, errors.New("error getting data from recharge_sites table from DB")
	}

	for i := 0; i < len(rechargeSites); i++ {
		rechargeSites[i].FileType = 117
	}

	return rechargeSites, nil
}

func GetRechargeData(v *Setup) (rechargeData []RechargeSiteData, err error) {
	rechargeQuery := "SELECT site_id, DATE_TRUNC('month', recarge_amount.rch_date) AS rch_month, SUM(value) AS value FROM sp.recarge_amount GROUP BY site_id, DATE_TRUNC('month',recarge_amount.rch_date);"

	if err := v.PgDb.Select(&rechargeData, rechargeQuery); err != nil {
		return rechargeData, errors.New("error getting data from recharge_sites table from DB")
	}

	return rechargeData, nil
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
			if sites[i].id == data[j].SiteId {
				sites[i].RchData = append(sites[i].RchData, data[j])
			}
		}
	}

	return sites, nil
}
