package database

import (
	"errors"
	"fmt"
	"time"
)

type Retirement struct {
	Id        int       `db:"id"`
	Area      float64   `db:"area"`
	SubArea   string    `db:"subarea"`
	StartDate time.Time `db:"start_date"`
	EndDate   time.Time `db:"end_date"`
	ProjPump  float64   `db:"proj_pumping"`
	RWells    RetirementWells
}

func (r Retirement) IsExpired(year int) bool {
	return r.EndDate.Before(time.Date(year, time.July, 1, 0, 0, 0, 0, time.UTC))
}

func (r Retirement) ToSqliteWelResults(year int) SqliteWelResults {
	results := make(SqliteWelResults, len(r.RWells))
	pPump := ((r.ProjPump / float64(len(r.RWells))) * r.Area) / 12.0
	for _, w := range r.RWells {
		for i := 0; i < 3; i++ {
			mnth := time.June
			pmp := pPump * 0.85 * 0.2
			switch i {
			case 1:
				{
					mnth = time.July
					pmp = pPump * 0.85 * 0.4
				}
			case 2:
				{
					mnth = time.August
					pmp = pPump * 0.85 * 0.4
				}
			}

			sqlWelResult := SqliteWelResult{
				WellId:   w.Well,
				CellNode: w.Node,
				Dt:       time.Date(year, mnth, 1, 0, 0, 0, 0, time.UTC),
				FileType: 221,
				Result:   pmp,
			}

			results = append(results, sqlWelResult)
		}
	}

	return results
}

type Retirements []Retirement

type RetirementWell struct {
	RId  int `db:"retirement"`
	Well int `db:"well"`
	Node int `db:"node"`
}

type RetirementWells []RetirementWell

func GetSingleYearRetirements(v *Setup, year int) (retirements Retirements, err error) {
	// only for years 2014 through 2020
	if year < 2008 {
		return retirements, errors.New("year must be 2008 through 2020")
	}

	npRetirements := Retirements{}
	selectStmt := fmt.Sprintf("select retirements.id, st_area(geom)/43560 area, retirements.subarea, start_date, end_date, proj_pumping from np.t%[1]d_dry inner join np.retirements on t%[1]d_dry.retirement = retirements.id inner join np.retirement_pumping on retirements.subarea = retirement_pumping.subarea where retirements.id > 1 and sw = false and gw = true and permanent = false;", year)
	if err := v.PgDb.Select(&npRetirements, selectStmt); err != nil {
		return retirements, err
	}

	spRetirements := Retirements{}
	selectStmt = fmt.Sprintf("select retirements.id, st_area(geom)/43560 area, retirements.subarea, start_date, end_date, proj_pumping from sp.t%[1]d_dry inner join sp.retirements on t%[1]d_dry.retirement = retirements.id inner join sp.retirement_pumping on retirements.subarea = retirement_pumping.subarea where retirements.id > 1 and sw = false and gw = true and permanent = false;", year)
	if err := v.PgDb.Select(&spRetirements, selectStmt); err != nil {
		return retirements, err
	}

	copy(retirements, append(npRetirements, spRetirements...))
	return retirements, nil
}

func GetRetirementWells(v *Setup, grid int) (rWells RetirementWells, err error) {

	npWells := RetirementWells{}
	query := fmt.Sprintf("select retirement, well, node from np.retirements_jct inner join np.npnrd_wells on well = npnrd_wells.wellid inner join model_cells mc on st_contains(mc.geom, npnrd_wells.geom) where mc.cell_type = %d;", grid)
	if err := v.PgDb.Select(&npWells, query); err != nil {
		return rWells, err
	}

	spWells := RetirementWells{}
	query = fmt.Sprintf("select retirement, well, node from sp.retirements_jct inner join sp.spnrd_wells on well = spnrd_wells.wellid inner join model_cells mc on st_contains(mc.geom, spnrd_wells.geom) where mc.cell_type = %d;", grid)
	if err := v.PgDb.Select(&spWells, query); err != nil {
		return rWells, err
	}

	copy(rWells, append(npWells, spWells...))
	return rWells, nil
}

func GetRetirements(v *Setup, grid int, rYears []int) (allRetirements map[int]Retirements, err error) {
	allRetirements = make(map[int]Retirements)

	allRetirementWells, err := GetRetirementWells(v, grid)
	if err != nil {
		return allRetirements, err
	}

	for _, rYear := range rYears {
		singleYrRetirements, err := GetSingleYearRetirements(v, rYear)
		if err != nil {
			return allRetirements, err
		}

		for i := 0; i < len(singleYrRetirements); i++ {
			for _, w := range allRetirementWells {
				if w.RId == singleYrRetirements[i].Id {
					singleYrRetirements[i].RWells = append(singleYrRetirements[i].RWells, w)
				}
			}
		}
	}

	return allRetirements, nil
}

func (rs *Retirements) SendRetirements(year int) (welResults SqliteWelResults, err error) {
	welResults = SqliteWelResults{}

	for i := 0; i < len(*rs); i++ {
		retirement := (*rs)[i]
		if retirement.IsExpired(year) {
			welResults = append(welResults, retirement.ToSqliteWelResults(year)...)
		}
	}

	return welResults, nil
}
