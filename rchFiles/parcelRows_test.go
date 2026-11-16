package rchFiles

import (
	"math"
	"testing"

	"github.com/Longitude103/ESC-Model/parcels"
)

// A 100 acre parcel that is 60 acres on soil 412 and 40 acres on soil 622, crossing three cells. All of the parcel's
// DP (from both soil rows) must be recharged, each cell getting the DP of its own soil row.
func TestCellParcelRow_ConservesParcelTotals(t *testing.T) {
	rows := []parcels.Parcel{
		{ParcelNo: 9, Nrd: "np", Area: 100, SoilArea: 60, SoilCode: 412, CoeffZone: 1, Dp: [12]float64{6: 6}},
		{ParcelNo: 9, Nrd: "np", Area: 100, SoilArea: 40, SoilCode: 622, CoeffZone: 1, Dp: [12]float64{6: 2}},
	}
	idx := indexParcels(rows)

	cells := []struct {
		soil    int
		irrArea float64
	}{{412, 35}, {412, 25}, {622, 40}}

	total := 0.0
	for _, c := range cells {
		row, area, matched := cellParcelRow(idx[parcelKey{id: 9, nrd: "np"}], c.soil, 1)
		if !matched {
			t.Fatalf("soil %d should match a parcel row", c.soil)
		}
		r, err := cellRCH(row.Ro[6], row.Dp[6], area, c.irrArea, 0.5, 0)
		if err != nil {
			t.Fatal(err)
		}
		total += r
	}

	if math.Abs(total-8) > 1e-9 {
		t.Errorf("recharged %f af, want all 8 af of parcel DP (the old first-row lookup gave 6)", total)
	}
}

func TestCellParcelRow_FallbackUsesWholeParcel(t *testing.T) {
	rows := []parcels.Parcel{
		{ParcelNo: 9, Area: 100, SoilArea: 60, SoilCode: 412, Dp: [12]float64{6: 6}},
		{ParcelNo: 9, Area: 100, SoilArea: 40, SoilCode: 622, Dp: [12]float64{6: 2}},
	}

	row, area, matched := cellParcelRow(rows, 999, 1)
	if matched || area != 100 || row.Dp[6] != 8 {
		t.Errorf("fallback should combine rows: matched=%t area=%f dp=%f", matched, area, row.Dp[6])
	}
}
