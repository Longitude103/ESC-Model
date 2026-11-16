package database

import (
	"errors"
	"fmt"
	"time"
)

type ExtRch struct {
	Node     int     `db:"node"`
	Size     float64 `db:"cell_size"`
	Yr       int     `db:"yr"`
	Mnth     int     `db:"mnth"`
	FileType int     `db:"file_type"`
	Rch      float64 `db:"rch"`
}

// GetExtRecharge is a function to return a slice of ExtRch filled with the ext_recharge data from the database and
// used in the model to create the recharge values for the external non-NRD areas.
func GetExtRecharge(v *Setup) (eRch []ExtRch, err error) {
	// ext_recharge only has data through its last year; later model years repeat that last year.
	lastDataYr, err := lastExtDataYear(v, "ext_recharge")
	if err != nil {
		return eRch, err
	}

	endYr := v.EYear
	if v.EYear > lastDataYr {
		endYr = lastDataYr
	}

	query := fmt.Sprintf("select yr, mnth, file_type, rch, node, st_area(mc.geom)/43560 cell_size from "+
		"ext_recharge inner join model_cells mc on st_contains(mc.geom, ext_recharge.geom) where yr >= %d "+
		"and yr <= %d and mc.cell_type = %d;", v.SYear, endYr, v.CellType())

	if err := v.PgDb.Select(&eRch, query); err != nil {
		return eRch, err
	}

	if v.EYear > lastDataYr {
		var extraRch []ExtRch
		additionalYrs := v.EYear - lastDataYr
		extraQuery := fmt.Sprintf("select yr, mnth, file_type, rch, node, st_area(mc.geom)/43560 cell_size from "+
			"ext_recharge inner join model_cells mc on st_contains(mc.geom, ext_recharge.geom) where yr = %d and mc.cell_type = %d;", lastDataYr, v.CellType())

		if err := v.PgDb.Select(&extraRch, extraQuery); err != nil {
			return eRch, errors.New("error getting data from ext_pumping table from DB")
		}

		for i := 0; i < additionalYrs; i++ {
			for j := 0; j < len(extraRch); j++ {
				extraRch[j].Yr = extraRch[j].Yr + 1
			}

			eRch = append(eRch, extraRch...)
		}
	}

	return eRch, nil
}

// Date is a method to return a formatted date object for the ExtRch struct using its components and setting the day
// to 1 and 0 hours in UTC.
func (e *ExtRch) Date() time.Time {
	return time.Date(e.Yr, time.Month(e.Mnth), 1, 0, 0, 0, 0, time.UTC)
}
