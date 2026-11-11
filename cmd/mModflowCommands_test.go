package cmd_test

import (
	"github.com/Longitude103/ESC-Model/cmd"
	"database/sql"
	"testing"
	"time"

	"github.com/Longitude103/ESC-Model/database"
)

func TestExcludeResults(t *testing.T) {
	mr5 := database.MfResults{CellNode: 4, CellSize: sql.NullFloat64{Float64: 40.0, Valid: true}, ResultDate: time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC), Rslt: 100.1, Rw: sql.NullInt64{Int64: 100, Valid: true}, Clm: sql.NullInt64{Int64: 2, Valid: true}, ConvertedValue: true}
	mr1 := database.MfResults{CellNode: 1, CellSize: sql.NullFloat64{Float64: 40.0, Valid: true}, ResultDate: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), Rslt: 100.1, Rw: sql.NullInt64{Int64: 100, Valid: true}, Clm: sql.NullInt64{Int64: 2, Valid: true}, ConvertedValue: true}
	mr2 := database.MfResults{CellNode: 2, CellSize: sql.NullFloat64{Float64: 40.0, Valid: true}, ResultDate: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), Rslt: 100.1, Rw: sql.NullInt64{Int64: 100, Valid: true}, Clm: sql.NullInt64{Int64: 2, Valid: true}, ConvertedValue: true}
	mr3 := database.MfResults{CellNode: 3, CellSize: sql.NullFloat64{Float64: 40.0, Valid: true}, ResultDate: time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC), Rslt: 100.1, Rw: sql.NullInt64{Int64: 100, Valid: true}, Clm: sql.NullInt64{Int64: 2, Valid: true}, ConvertedValue: true}
	mr4 := database.MfResults{CellNode: 4, CellSize: sql.NullFloat64{Float64: 40.0, Valid: true}, ResultDate: time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC), Rslt: 100.1, Rw: sql.NullInt64{Int64: 100, Valid: true}, Clm: sql.NullInt64{Int64: 2, Valid: true}, ConvertedValue: true}

	testMR := cmd.SliceMfResults{mr5, mr1, mr2, mr3, mr4}
	excludeTest := testMR.ExcludeResults([]int{2018, 2019, 2016})

	if len(excludeTest) == 0 {
		t.Error("didn't return any records, expecting 2")
	}

	for _, e := range excludeTest {
		if e.Year() == 2018 || e.Year() == 2019 || e.Year() == 2016 {
			t.Error("found a year that should have been excluded")
		}
	}
}
