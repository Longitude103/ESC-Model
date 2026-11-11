package database

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/Longitude103/ESC-Model/Utils"
)

type WellParcel struct {
	ParcelId int    `db:"parcel_id"`
	WellId   int    `db:"wellid"`
	Nrd      string `db:"nrd"`
	Yr       int    `db:"yr"`
}

type WellNode struct {
	WellId int            `db:"wellid"`
	RegCd  sql.NullString `db:"regcd"`
	Node   int            `db:"node"`
	Nrd    string         `db:"nrd"`
}

type SSWell struct {
	Id       int `db:"id"`
	WellName int `db:"wellname"`
	Rate     int `db:"defaultq"`
	Node     int `db:"node"`
	MVolume  [12]float64
}

type MIWell struct {
	WellId     int    `db:"id"`
	WellName   string `db:"wellname"`
	Rate       int    `db:"defaultq"`
	MuniWell   bool   `db:"muni_well"`
	IndustWell bool   `db:"indust_well"`
	Stop97     bool   `db:"stop_97"`
	Start97    bool   `db:"start_97"`
	Nrd        string `db:"nrd"`
	Node       int    `db:"node"`
	Pumping    []MIPumping
}

type MIPumping struct {
	WellId   int       `db:"well_id"`
	PumpDate time.Time `db:"dt"`
	Pump     float64   `db:"pumping"`
}

// GetWellParcels is a function that gets all the well parcel junction table values and creates one struct from them
// and also includes the year of the join as well as the nrd.
func GetWellParcels(v *Setup) ([]WellParcel, error) {
	if v.Post97 {
		return GetWellParcelsPost97(v)
	}

	const query = "select parcel_id, wellid, nrd, yr from public.alljct();"

	var wellParcels []WellParcel
	if err := v.PgDb.Select(&wellParcels, query); err != nil {
		fmt.Println("Err: ", err)
		return wellParcels, errors.New("error getting parcel wells from db function")
	}

	if v.AppDebug && len(wellParcels) > 50 {
		return wellParcels[:50], nil
	}

	return wellParcels, nil
}

// GetWellParcelsPost97 is a function that gets all the well parcel junction table values and creates one struct from them
// and also includes the year of the join as well as the nrd but replaces any GWO parcels with the 1997 GWO parcels.
func GetWellParcelsPost97(v *Setup) ([]WellParcel, error) {
	pre97queryFormatString := `select parcel_id, wellid, 'np' nrd, %d yr from np.t%d_jct union all
								select parcel_id, wellid, 'sp' nrd, %d yr from sp.t%d_jct;`

	var wellParcels97GWO []WellParcel
	// query to get 1997 GWO parcel well relationships
	gwo97Query := `select j.parcel_id + 30000 as parcel_id, j.wellid, 'np' nrd, 1997 yr from np.t1997_jct j inner join (select parcel_id from np.t1997_irr where sw = false and gw = true) gwo on gwo.parcel_id = j.parcel_id 
					union all 
				   select j.parcel_id + 30000 as parcel_id, j.wellid, 'sp' nrd, 1997 yr from sp.t1997_jct j inner join (select parcel_id from sp.t1997_irr where sw = false and gw = true) gwo on gwo.parcel_id = j.parcel_id;`

	if err := v.PgDb.Select(&wellParcels97GWO, gwo97Query); err != nil {
		fmt.Println("Err: ", err)
		return nil, errors.New("error getting 97 Ground water only parcel wells from db function")
	}

	var allWellParcels []WellParcel
	for yr := v.SYear; yr < v.EYear+1; yr++ {
		var qry string
		var wellParcels []WellParcel
		if yr < 1998 {
			qry = fmt.Sprintf(pre97queryFormatString, yr, yr, yr, yr)

			if err := v.PgDb.Select(&wellParcels, qry); err != nil {
				fmt.Println("Err: ", err)
				return wellParcels, errors.New("error getting parcel wells from db function")
			}
		} else {
			coMingledQuery := fmt.Sprintf(`select j.parcel_id, j.wellid, 'np' nrd, %d yr from np.t%d_jct j inner join (select parcel_id from np.t%d_irr where sw = true and gw = true) gwo on gwo.parcel_id = j.parcel_id
													union all
 												  select j.parcel_id, j.wellid, 'sp' nrd, %d yr from sp.t%d_jct j inner join (select parcel_id from sp.t%d_irr where sw = true and gw = true) gwo on gwo.parcel_id = j.parcel_id;`, yr, yr, yr, yr, yr, yr)

			if err := v.PgDb.Select(&wellParcels, coMingledQuery); err != nil {
				fmt.Println("Err: ", err)
				return wellParcels, errors.New("error getting parcel wells from db function")
			}

			for i := 0; i < len(wellParcels97GWO); i++ {
				modWellParcel := wellParcels97GWO[i]
				modWellParcel.Yr = yr

				wellParcels = append(wellParcels, modWellParcel)
			}
		}

		allWellParcels = append(allWellParcels, wellParcels...)
	}

	return allWellParcels, nil
}

// GetWellNode is a function that gets the wellid, regno and node number of the well so that we can add a location to
// the well when it is written out along with the nrd.
func GetWellNode(v *Setup) (wellNodes []WellNode, err error) {
	query := fmt.Sprintf("select wellid, regcd, node, 'np' nrd from np.npnrd_wells nw inner join model_cells mc "+
		"on st_contains(mc.geom, nw.geom) where mc.cell_type = %d union all select wellid, regcd, node, 'sp' nrd from sp.spnrd_wells sw "+
		"inner join model_cells mc on st_contains(mc.geom, sw.geom) where mc.cell_type = %d", v.CellType(), v.CellType())

	if err := v.PgDb.Select(&wellNodes, query); err != nil {
		return wellNodes, errors.New("error getting well node locations from DB")
	}

	if v.AppDebug && len(wellNodes) > 50 {
		return wellNodes[:50], nil
	}

	return wellNodes, nil
}

// GetSSWells is a function that gets the data from the postgres DB and returns a slice of SSWell and also includes a call
// to the SSWell.monthlyVolume() method to set the monthly data from the annual data that is in the database.
func GetSSWells(v *Setup) (ssWells []SSWell, err error) {
	ssQuery := fmt.Sprintf("select ss_wells.id, wellname, defaultq, node from ss_wells inner join model_cells mc on "+
		"st_contains(mc.geom, st_translate(ss_wells.geom, 20, 20)) where mc.cell_type = %d;", v.CellType())

	if err := v.PgDb.Select(&ssWells, ssQuery); err != nil {
		return ssWells, errors.New("error getting steady state wells from DB")
	}

	for i := 0; i < len(ssWells); i++ {
		if err := ssWells[i].monthlyVolume(); err != nil {
			return ssWells, errors.New("error setting monthly volumes")
		}
	}

	if v.AppDebug {
		return ssWells[:50], nil
	}

	return ssWells, nil
}

// monthlyVolume is a method of SSWell that calculates the monthly volume of pumping from the rate that is included in the
// database records. It turns the value positive to make it uniform with the other results.
func (s *SSWell) monthlyVolume() (err error) {
	annVolume := -1.0 * float64(s.Rate) * 365.25 / 43560

	for i := 0; i < 12; i++ {
		s.MVolume[i] = (annVolume / 12)
	}

	return nil
}

//go:embed sql/wells.sql
var queries string

func GetMIWells(v *Setup) (miWells []MIWell, err error) {
	formattedQueries := Utils.SplitQueries(queries)

	miQuery := fmt.Sprintf(formattedQueries[0], v.CellType())

	if err = v.PgDb.Select(&miWells, miQuery); err != nil {
		return miWells, errors.New("error getting Municipal and industrial wells: " + err.Error())
	}

	// use setup to get bounds of pumping (earliest is 1997)
	var miPumpQuery string
	if v.Post97 {
		miPumpQuery = fmt.Sprintf(formattedQueries[1], 1997, 1997)
	} else {
		miPumpQuery = fmt.Sprintf(formattedQueries[1], v.SYear, v.EYear)
	}

	var miPump []MIPumping
	if err = v.PgDb.Select(&miPump, miPumpQuery); err != nil {
		return miWells, errors.New("error getting Pumping for M&I wells: " + err.Error())
	}

	for i := 0; i < len(miWells); i++ {
		for _, p := range miPump {
			if p.WellId == miWells[i].WellId {
				miWells[i].Pumping = append(miWells[i].Pumping, p)
			}
		}
	}

	return
}

func (w *MIWell) MIFileType() int {
	switch w.Nrd {
	case "np":
		if w.MuniWell {
			return 210
		} else if w.IndustWell {
			return 211
		} else {
			return 212
		}
	case "sp":
		if w.MuniWell {
			return 219
		} else if w.IndustWell {
			return 220
		} else {
			return 212
		}
	default:
		return 212
	}
}
