package conveyLoss

import (
	"database/sql"
	"testing"
	"time"
)

func Test_getDiversions(t *testing.T) {
	v := dbConnection(t)

	v.SYear = 1953
	v.EYear = 1953

	divs, _, err := getDiversions(v)
	if err != nil {
		t.Error("Get Diversions errored")
	}

	for _, div := range divs {
		if div.CanalId == 26 {
			v.Logger.Debugf("Laramie Div: %+v", div)
		}

		if div.CanalId == 272 {
			v.Logger.Debugf("Laramie WY Div: %+v", div)
		}

		if div.CanalId == 13 {
			v.Logger.Debugf("Mitchell Div: %+v", div)
		}

		if div.CanalId == 32 {
			v.Logger.Debugf("Gering Div: %+v", div)
		}

		if div.CanalId == 15 {
			v.Logger.Debugf("Enterprise Div: %+v", div)
		}

		if div.CanalId == 29 {
			v.Logger.Debugf("Minatare Div: %+v", div)
		}
	}

	// if divs[1].DivAmount.Float64 != 12797.0 {
	// 	t.Errorf("Wrong amount being queried: Should be 12797, got %f", divs[1].DivAmount.Float64)
	// }

}

func Test_getDiversionsSS(t *testing.T) {
	v := dbConnection(t)
	_ = v.SetYears(1895, 1905)
	v.SteadyState = true

	divs, _, err := getDiversions(v)
	if err != nil {
		t.Error("Get Diversions errored")
	}

	for _, d := range divs {
		if d.DivDate.Time.Year() > 1952 || d.DivDate.Time.Year() < 1895 {
			t.Errorf("Div Dates are wrong: %+v", d)
		}

		if d.CanalId == 0 {
			t.Errorf("Bad Canal Id: %+v", d)
		}
	}
}

func Test_getEFDiversions(t *testing.T) {
	v := dbConnection(t)
	if err := v.SetYears(2011, 2020); err != nil {
		t.Error("Failed to set years")
	}

	_, div, err := getDiversions(v)
	if err != nil {
		t.Error("Error in getDiversions: ", err)
	}

	want := 48
	got := len(div)
	if want != got {
		t.Errorf("getDiversions should have returned %d diversion records but got %d records", want, got)
	}
}

func TestEfPeriod_GetYearAndMonths(t *testing.T) {
	ef := efPeriod{
		CanalId:    1,
		StartDate:  sql.NullTime{Valid: true, Time: time.Date(2016, 4, 1, 0, 0, 0, 0, time.UTC)},
		EndDate:    sql.NullTime{Valid: true, Time: time.Date(2016, 6, 6, 0, 0, 0, 0, time.UTC)},
		LossPercet: sql.NullFloat64{Valid: true, Float64: 0.56},
	}

	y, months := ef.GetYearAndMonths()

	if y != 2016 {
		t.Errorf("Was expecting year 2016, got year %d", y)
	}

	if len(months) != 3 {
		t.Errorf("was expecting a return of 3 months, but got %d months", len(months))
	}

	for _, m := range months {
		if m < 4 && m > 6 {
			t.Errorf("months should be 4, 5, or 6, but got %d", m)
		}
	}
}

func Test_FindDiversion(t *testing.T) {
	diversion := Diversion{
		CanalId:   5,
		DivDate:   sql.NullTime{Time: time.Date(2011, 4, 5, 0, 0, 0, 0, time.UTC), Valid: true},
		DivAmount: sql.NullFloat64{Float64: 100.1, Valid: true},
	}

	diversion2 := Diversion{
		CanalId:   5,
		DivDate:   sql.NullTime{Time: time.Date(2011, 3, 25, 0, 0, 0, 0, time.UTC), Valid: true},
		DivAmount: sql.NullFloat64{Float64: 100.1, Valid: true},
	}

	diversion3 := Diversion{
		CanalId:   5,
		DivDate:   sql.NullTime{Time: time.Date(2011, 4, 1, 0, 0, 0, 0, time.UTC), Valid: true},
		DivAmount: sql.NullFloat64{Float64: 100.1, Valid: true},
	}

	diversion4 := Diversion{
		CanalId:   5,
		DivDate:   sql.NullTime{Time: time.Date(2011, 6, 2, 0, 0, 0, 0, time.UTC), Valid: true},
		DivAmount: sql.NullFloat64{Float64: 100.1, Valid: true},
	}

	p1 := efPeriod{
		CanalId:    5,
		StartDate:  sql.NullTime{Time: time.Date(2011, 4, 1, 0, 0, 0, 0, time.UTC), Valid: true},
		EndDate:    sql.NullTime{Time: time.Date(2011, 4, 25, 0, 0, 0, 0, time.UTC), Valid: true},
		LossPercet: sql.NullFloat64{Float64: 0.45, Valid: true},
	}

	p2 := efPeriod{
		CanalId:    5,
		StartDate:  sql.NullTime{Time: time.Date(2011, 5, 3, 0, 0, 0, 0, time.UTC), Valid: true},
		EndDate:    sql.NullTime{Time: time.Date(2011, 6, 2, 0, 0, 0, 0, time.UTC), Valid: true},
		LossPercet: sql.NullFloat64{Float64: 0.45, Valid: true},
	}

	periods := []efPeriod{p1, p2}

	if !findDiversion(diversion, periods) {
		t.Error("Should have produced true.")
	}

	if findDiversion(diversion2, periods) {
		t.Error("Should have produced false.")
	}

	if !findDiversion(diversion3, periods) {
		t.Error("Should have produced true.")
	}

	if !findDiversion(diversion4, periods) {
		t.Error("Should have produced true.")
	}
}
