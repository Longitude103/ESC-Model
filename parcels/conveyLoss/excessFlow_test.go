package conveyLoss

import (
	"database/sql"
	"reflect"
	"testing"
	"time"
)

func efp(canal int, start, end time.Time, loss float64) efPeriod {
	return efPeriod{CanalId: canal, StartDate: sql.NullTime{Valid: true, Time: start}, EndDate: sql.NullTime{Valid: true, Time: end},
		LossPercet: sql.NullFloat64{Valid: true, Float64: loss}}
}

func d(y int, m time.Month, day int) time.Time { return time.Date(y, m, day, 0, 0, 0, 0, time.UTC) }

func TestGetYearAndMonths_Edges(t *testing.T) {
	// starts on the last day of April: April must be included
	e := efp(1, d(2016, 4, 30), d(2016, 5, 10), 0.5)
	if _, m := e.GetYearAndMonths(); !reflect.DeepEqual(m, []int{4, 5}) {
		t.Errorf("got months %v, want [4 5]", m)
	}

	// ends on the last day of June: July must not be added
	e = efp(1, d(2016, 4, 1), d(2016, 6, 30), 0.5)
	if _, m := e.GetYearAndMonths(); !reflect.DeepEqual(m, []int{4, 5, 6}) {
		t.Errorf("got months %v, want [4 5 6]", m)
	}
}

func TestMatchEFPeriods_NoDoubleCount(t *testing.T) {
	june := Diversion{CanalId: 7, DivDate: sql.NullTime{Valid: true, Time: d(2016, 6, 1)}, DivAmount: sql.NullFloat64{Valid: true, Float64: 100}}
	periods := []efPeriod{
		efp(7, d(2016, 6, 2), d(2016, 6, 8), 0.4),
		efp(7, d(2016, 6, 20), d(2016, 6, 25), 0.4),
		efp(8, d(2016, 6, 1), d(2016, 6, 30), 0.9),
	}

	got := matchEFPeriods([]Diversion{june}, periods)
	if len(got) != 1 || got[0].LossPercent.Float64 != 0.4 {
		t.Errorf("June excess flow should be matched once to canal 7, got %+v", got)
	}
}
