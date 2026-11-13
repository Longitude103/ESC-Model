package database

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/Longitude103/ESC-Model/Utils"
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
	// ext_pumping only has data through its last year; later model years repeat that last year.
	lastDataYr, err := lastExtDataYear(v, "ext_pumping")
	if err != nil {
		return extWells, err
	}

	endYr := v.EYear
	if v.EYear > lastDataYr {
		endYr = lastDataYr
	}

	extQuery := fmt.Sprintf("select yr, mnth, file_type, pmp, node from ext_pumping inner join model_cells mc on "+
		"st_contains(mc.geom, ext_pumping.geom) where yr >= %d and yr <= %d and mc.cell_type = %d;", v.SYear, endYr, v.CellType())

	if err := v.PgDb.Select(&extWells, extQuery); err != nil {
		return extWells, errors.New("error getting data from ext_pumping table from DB")
	}

	if v.AppDebug {
		return extWells[:50], nil
	}

	if v.EYear > lastDataYr {
		var extraWells []ExtWell
		additionalYrs := v.EYear - lastDataYr
		extraQuery := fmt.Sprintf("select yr, mnth, file_type, pmp, node from ext_pumping inner join model_cells mc on "+
			"st_contains(mc.geom, ext_pumping.geom) where yr = %d and mc.cell_type = %d;", lastDataYr, v.CellType())

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
