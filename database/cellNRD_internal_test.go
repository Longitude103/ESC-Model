package database

import (
	"database/sql"
	"math"
	"testing"
)

func nf(f float64) sql.NullFloat64 { return sql.NullFloat64{Float64: f, Valid: true} }

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestVegAreasByFileType(t *testing.T) {
	tests := []struct {
		name string
		cell CellIntersect
		want map[int]float64
	}{
		{"all NP", CellIntersect{CellArea: 40, NpIrrArea: nf(10), NpFrac: 1},
			map[int]float64{NaturalVegNP: 30}},
		{"boundary cell split by area", CellIntersect{CellArea: 40, NpFrac: 0.25, SpFrac: 0.75},
			map[int]float64{NaturalVegNP: 10, NaturalVegSP: 30}},
		{"part outside both NRDs", CellIntersect{CellArea: 40, SpIrrArea: nf(20), SpFrac: 0.5},
			map[int]float64{NaturalVegSP: 10, NaturalVegOther: 10}},
		{"not in any zone", CellIntersect{CellArea: 40},
			map[int]float64{NaturalVegOther: 40}},
		{"no natural veg", CellIntersect{CellArea: 40, NpIrrArea: nf(45), NpFrac: 1}, map[int]float64{}},
		// SP cell that lost a 15 acre NP parcel (parcel's NRD differs from most of the cell): base veg 10 split
		// 80/20, plus the removed 15 acres all NP
		{"removed parcel goes to its NRD", CellIntersect{CellArea: 40, SpIrrArea: nf(15), NpRemoved: 15, NpFrac: 0.2, SpFrac: 0.8},
			map[int]float64{NaturalVegNP: 2 + 15, NaturalVegSP: 8}},
		{"fractions over 1 are scaled", CellIntersect{CellArea: 40, NpFrac: 0.6, SpFrac: 0.6},
			map[int]float64{NaturalVegNP: 20, NaturalVegSP: 20}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cell.VegAreasByFileType()
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			var sum float64
			for ft, a := range tt.want {
				if !near(got[ft], a) {
					t.Errorf("file type %d: got %f, want %f", ft, got[ft], a)
				}
				sum += got[ft]
			}
			if !near(sum, tt.cell.VegArea()) {
				t.Errorf("areas add to %f, want VegArea %f", sum, tt.cell.VegArea())
			}
		})
	}
}

func TestSetRemovedAreas(t *testing.T) {
	base := []CellIntersect{{Node: 1, Soil: 2, CZone: 3, NpIrrArea: nf(10), SpDryArea: nf(5), SpIrrArea: nf(4)}}
	cells := []CellIntersect{{Node: 1, Soil: 2, CZone: 3, NpIrrArea: nf(2), SpIrrArea: nf(6)},
		{Node: 9, Soil: 2, CZone: 3}}

	setRemovedAreas(cells, base)

	if !near(cells[0].NpRemoved, 8) || !near(cells[0].SpRemoved, 3) {
		t.Errorf("got NP %f SP %f removed, want 8 and 3", cells[0].NpRemoved, cells[0].SpRemoved)
	}
	if cells[1].NpRemoved != 0 || cells[1].SpRemoved != 0 {
		t.Errorf("cell not in base should have nothing removed, got %+v", cells[1])
	}
}

func TestCellNRDFractions(t *testing.T) {
	got := cellNRDFractions([]cellNRDArea{
		{Node: 1, Nrd: "NP", Frac: 0.3}, {Node: 1, Nrd: "sp", Frac: 0.7},
		{Node: 2, Nrd: "np", Frac: 0.5}, {Node: 2, Nrd: "np", Frac: 0.25}, {Node: 2, Nrd: "xx", Frac: 0.25},
		{Node: 3, Nrd: "np", Frac: 0.8}, {Node: 3, Nrd: "sp", Frac: 0.4},
	})

	want := map[int]CellNRDFraction{1: {0.3, 0.7}, 2: {0.75, 0}, 3: {0.8 / 1.2, 0.4 / 1.2}}
	for n, w := range want {
		if !near(got[n].NP, w.NP) || !near(got[n].SP, w.SP) {
			t.Errorf("node %d: got %+v, want %+v", n, got[n], w)
		}
	}
}
