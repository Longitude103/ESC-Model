package rchFiles

import (
	"database/sql"
	"math"
	"testing"

	"github.com/Longitude103/ESC-Model/database"
)

func TestNaturalVegResultsSplitByArea(t *testing.T) {
	cell := database.CellIntersect{Node: 5, CellArea: 40, NpIrrArea: sql.NullFloat64{Float64: 20, Valid: true},
		NpFrac: 0.25, SpFrac: 0.75}
	var perAcre [12]float64
	perAcre[6] = 0.1

	results := naturalVegResults(cell, 2000, perAcre)
	if len(results) != 2 {
		t.Fatalf("want one NP and one SP result for July, got %+v", results)
	}

	got := map[int]float64{}
	for _, r := range results {
		got[r.FileType] = r.Result
		if r.Node != 5 || r.Dt.Year() != 2000 || r.Dt.Month() != 7 || r.Size != 40 {
			t.Errorf("unexpected result %+v", r)
		}
	}

	if math.Abs(got[database.NaturalVegNP]-0.5) > 1e-9 || math.Abs(got[database.NaturalVegSP]-1.5) > 1e-9 {
		t.Errorf("got %v, want NP 0.5 and SP 1.5", got)
	}
}

// the split relies on the recharge being linear in area, so one acre times the area must equal the old calculation
func TestNaturalVegLinearInArea(t *testing.T) {
	const area = 37.5
	_, rch1 := calcRo(1.2, 0.3, 0.4, 1, 0.9, 0.6, 0.5)
	dp1 := calcDp(2.1, 0.2, rch1, 0.4, 1, 1.1)

	_, rchA := calcRo(1.2, 0.3, 0.4, area, 0.9, 0.6, 0.5)
	dpA := calcDp(2.1, 0.2, rchA, 0.4, area, 1.1)

	if math.Abs(dp1*area-dpA) > 1e-9 {
		t.Errorf("per acre %f x %f = %f, full area gives %f", dp1, area, dp1*area, dpA)
	}
}
