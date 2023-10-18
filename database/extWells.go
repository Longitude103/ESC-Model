package database

import (
	"errors"
	"fmt"
	"math"
	"time"

	"bitbucket.org/longitude-103/wwumm-esc/Utils"
)

type ExtWell struct {
	Yr       int     `db:"yr"`
	Mnth     int     `db:"mnth"`
	FileType int     `db:"file_type"`
	Pumping  float64 `db:"pmp"`
	Node     int     `db:"node"`
}

// GetExternalWells is a function to query the external pumping from the database and returns a slice of ExtWell as well
// as includes handling the debug mode.
func GetExternalWells(v *Setup) (extWells []ExtWell, err error) {
	endYr := v.EYear
	if v.EYear > 2014 {
		endYr = 2014
	}

	extQuery := fmt.Sprintf("select yr, mnth, file_type, pmp, node from ext_pumping inner join model_cells mc on "+
		"st_contains(mc.geom, ext_pumping.geom) where yr >= %d and yr <= %d and mc.cell_type = %d;", v.SYear, endYr, v.CellType())

	if err := v.PgDb.Select(&extWells, extQuery); err != nil {
		return extWells, errors.New("error getting data from ext_pumping table from DB")
	}

	if v.AppDebug {
		return extWells[:50], nil
	}

	// actual data is only through 2014, might remove this if the data is updated.
	if v.EYear > 2014 {
		var extraWells []ExtWell
		additionalYrs := v.EYear - 2014
		extraQuery := fmt.Sprintf("select yr, mnth, file_type, pmp, node from ext_pumping inner join model_cells mc on "+
			"st_contains(mc.geom, ext_pumping.geom) where yr = 2014 and mc.cell_type = %d;", v.CellType())

		if err := v.PgDb.Select(&extraWells, extraQuery); err != nil {
			return extWells, errors.New("error getting data from ext_pumping table from DB")
		}

		for i := 0; i < additionalYrs; i++ {
			for j := 0; j < len(extraWells); j++ {
				extraWells[j].Yr = extraWells[j].Yr + 1
			}

			extWells = append(extWells, extraWells...)
		}
	}

	return extWells, nil
}

// Date returns the formatted date of the ExtWell struct with a 1 for the day and zero hour in UTC.
func (w *ExtWell) Date() time.Time {
	return time.Date(w.Yr, time.Month(w.Mnth), 1, 0, 0, 0, 0, time.UTC)
}

// Pmp is a method that returns the correct pumping values for each value
func (w *ExtWell) Pmp() float64 {
	// these two are rate amounts
	if w.FileType == 214 || w.FileType == 215 {
		pumpMonth := Utils.TimeExt{T: w.Date()}
		return math.Abs(w.Pumping) * float64(pumpMonth.DaysInMonth()) / 43560
	}

	// the remaining are acre-feet but need to be positive
	return math.Abs(w.Pumping)
}
