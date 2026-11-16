package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Retirement struct {
	Id        int          `db:"id"`
	Area      float64      `db:"area"`
	SubArea   string       `db:"subarea"`
	StartDate time.Time    `db:"start_date"`
	EndDate   sql.NullTime `db:"end_date"`
	ProjPump  float64      `db:"proj_pumping"`
	Nrd       string
	RWells    RetirementWells
}

func (r Retirement) IsExpired(year int) bool {
	if !r.EndDate.Valid {
		return false
	}

	return r.EndDate.Time.Before(time.Date(year, time.July, 1, 0, 0, 0, 0, time.UTC))
}

func (r Retirement) ToSqliteWelResults(year int, fullCorn bool) SqliteWelResults {
	if r.RWells == nil || len(r.RWells) == 0 {
		return nil
	}

	var results SqliteWelResults
	pmpAmount := r.ProjPump
	if fullCorn {
		pmpAmount = 15
	}

	pPump := ((pmpAmount / float64(len(r.RWells))) * r.Area) / 12.0
	for _, w := range r.RWells {
		for i := 0; i < 3; i++ {
			var mnth time.Month
			var pmp float64
			switch i {
			case 0:
				mnth = time.June
				pmp = pPump * 0.85 * 0.2
			case 1:
				mnth = time.July
				pmp = pPump * 0.85 * 0.4
			case 2:
				mnth = time.August
				pmp = pPump * 0.85 * 0.4
			}

			var ft int
			switch r.Nrd {
			case "NP":
				ft = 221
			default:
				ft = 222
			}

			sqlWelResult := SqliteWelResult{
				WellId:   w.Well,
				CellNode: w.Node,
				Dt:       time.Date(year, mnth, 1, 0, 0, 0, 0, time.UTC),
				FileType: ft,
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
	Nrd  string
}

type RetirementWells []RetirementWell

func GetSingleYearRetirements(v *Setup, year int) (retirements Retirements, err error) {
	// retirements start in 1998
	if year < 1998 {
		return retirements, errors.New("year must be 1998 or later")
	}

	npRetirements := Retirements{}
	var selectStmt string
	if v.PermRetire {
		selectStmt = fmt.Sprintf("select retirements.id, st_area(geom)/43560 area, retirements.subarea, start_date, end_date, proj_pumping from np.t%[1]d_dry inner join np.retirements on t%[1]d_dry.retirement = retirements.id inner join np.retirement_pumping on retirements.subarea = retirement_pumping.subarea where retirements.id > 1 and sw = false and gw = true;", year)
	} else {
		selectStmt = fmt.Sprintf("select retirements.id, st_area(geom)/43560 area, retirements.subarea, start_date, end_date, proj_pumping from np.t%[1]d_dry inner join np.retirements on t%[1]d_dry.retirement = retirements.id inner join np.retirement_pumping on retirements.subarea = retirement_pumping.subarea where retirements.id > 1 and sw = false and gw = true and permanent = false;", year)
	}

	if err := v.PgDb.Select(&npRetirements, selectStmt); err != nil {
		return retirements, err
	}

	for i := 0; i < len(npRetirements); i++ {
		npRetirements[i].Nrd = "NP"
	}

	spRetirements := Retirements{}
	if v.PermRetire {
		selectStmt = fmt.Sprintf("select retirements.id, st_area(geom)/43560 area, retirements.subarea, start_date, end_date, proj_pumping from sp.t%[1]d_dry inner join sp.retirements on t%[1]d_dry.retirement = retirements.id inner join sp.retirement_pumping on retirements.subarea = retirement_pumping.subarea where retirements.id > 1 and sw = false and gw = true;", year)
	} else {
		selectStmt = fmt.Sprintf("select retirements.id, st_area(geom)/43560 area, retirements.subarea, start_date, end_date, proj_pumping from sp.t%[1]d_dry inner join sp.retirements on t%[1]d_dry.retirement = retirements.id inner join sp.retirement_pumping on retirements.subarea = retirement_pumping.subarea where retirements.id > 1 and sw = false and gw = true and permanent = false;", year)
	}

	if err := v.PgDb.Select(&spRetirements, selectStmt); err != nil {
		return retirements, err
	}

	for i := 0; i < len(spRetirements); i++ {
		spRetirements[i].Nrd = "SP"
	}
	retirements = append(npRetirements, spRetirements...)
	return retirements, nil
}

func GetRetirementWells(v *Setup) (rWells RetirementWells, err error) {
	npWells := RetirementWells{}
	query := fmt.Sprintf("select retirement, well, node from np.retirements_jct inner join np.npnrd_wells on well = npnrd_wells.wellid inner join model_cells mc on st_contains(mc.geom, npnrd_wells.geom) where mc.cell_type = %d;", v.CellType())
	if err := v.PgDb.Select(&npWells, query); err != nil {
		return rWells, err
	}

	for i := 0; i < len(npWells); i++ {
		npWells[i].Nrd = "NP"
	}

	spWells := RetirementWells{}
	query = fmt.Sprintf("select retirement, well, node from sp.retirements_jct inner join sp.spnrd_wells on well = spnrd_wells.wellid inner join model_cells mc on st_contains(mc.geom, spnrd_wells.geom) where mc.cell_type = %d;", v.CellType())
	if err := v.PgDb.Select(&spWells, query); err != nil {
		return rWells, err
	}

	for i := 0; i < len(spWells); i++ {
		spWells[i].Nrd = "SP"
	}

	rWells = append(npWells, spWells...)
	return rWells, nil
}

func GetRetirements(v *Setup, rYears []int) (allRetirements map[int]Retirements, err error) {
	allRetirements = make(map[int]Retirements)
	allRetirementWells, err := GetRetirementWells(v)

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

		allRetirements[rYear] = singleYrRetirements
	}

	return allRetirements, nil
}

func GetOneYearRetirements(v *Setup, year int) (Retirements, error) {
	allRetirementWells, err := GetRetirementWells(v)
	if err != nil {
		return Retirements{}, err
	}

	singleYrRetirements, err := GetSingleYearRetirements(v, year)
	if err != nil {
		if err.Error() == "year must be 1998 or later" {
			v.Logger.Errorf("Year %d is not a valid retirement year, skipped", year)
			return Retirements{}, nil
		} else {
			return Retirements{}, err
		}
	}

	for i := 0; i < len(singleYrRetirements); i++ {
		for _, w := range allRetirementWells {
			if w.RId == singleYrRetirements[i].Id {
				singleYrRetirements[i].RWells = append(singleYrRetirements[i].RWells, w)
			}
		}
	}

	return singleYrRetirements, nil
}

func (rs *Retirements) SendRetirements(year int, fullCorn bool) (welResults SqliteWelResults, err error) {
	welResults = SqliteWelResults{}

	for i := 0; i < len(*rs); i++ {
		retirement := (*rs)[i]
		if retirement.RWells == nil || len(retirement.RWells) == 0 {
			continue
		}

		if retirement.IsExpired(year) {
			welResults = append(welResults, retirement.ToSqliteWelResults(year, fullCorn)...)
		}
	}

	return welResults, nil
}
