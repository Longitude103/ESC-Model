package conveyLoss

import (
	"database/sql"
	"math"
	"testing"
	"time"
)

func mDiv(canal int, month time.Month, cfsDays float64) Diversion {
	return Diversion{CanalId: canal, DivDate: sql.NullTime{Valid: true, Time: time.Date(1990, month, 1, 0, 0, 0, 0, time.UTC)},
		DivAmount: sql.NullFloat64{Valid: true, Float64: cfsDays}}
}

func TestFixMitchell(t *testing.T) {
	gering := getGeringDivs(Diversions{mDiv(32, time.June, 100)})

	// June: (1000 * 0.6 - 100) day-cfs * 1.9835 / 500 ac
	got := fixMitchell(mDiv(13, time.June, 1000), gering, 0.6, 500)
	if want := 500 * 1.9835 / 500; math.Abs(got-want) > 1e-9 {
		t.Errorf("June delivery %f, want %f (must include the day-cfs to acre-feet factor)", got, want)
	}

	// July: no Gering record, so nothing is removed (it used to return 0)
	got = fixMitchell(mDiv(13, time.July, 1000), gering, 0.6, 500)
	if want := 600 * 1.9835 / 500; math.Abs(got-want) > 1e-9 {
		t.Errorf("July delivery %f, want %f", got, want)
	}

	// Gering larger than Mitchell's delivered water never gives negative delivery
	if got = fixMitchell(mDiv(13, time.June, 100), gering, 0.6, 500); got != 0 {
		t.Errorf("delivery %f, want 0", got)
	}
}
