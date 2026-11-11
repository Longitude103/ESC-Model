package parcels_test

import (
	"github.com/Longitude103/ESC-Model/Utils"
	"github.com/Longitude103/ESC-Model/fileio"
	"github.com/Longitude103/ESC-Model/parcels"
	"testing"
)

func TestParcel_pValues(t *testing.T) {
	c1 := [12]float64{0, 0, 0, 0, 1.5, 3.5, 8, 7.5, 3, 0, 0}
	c2 := [12]float64{0, 0, 0, 0, 2, 4, 8.5, 8, 3.5, 0, 0}
	c3 := [12]float64{0, 0, 0, 0, 2.5, 4.5, 9, 8.5, 4, 0, 0}
	c4 := [12]float64{0, 0, 0, 0, 3, 5, 9.5, 9, 4.5, 0, 0}
	crops := [4][12]float64{c1, c2, c3, c4}

	cCoverage := [4]float64{0.35, 0.35, 0.3, 0.0}

	v := parcels.PValues(p1.Nir, crops, cCoverage, 0.5, 10)

	v1 := Utils.RoundTo(v[4], 2)
	v2 := Utils.RoundTo(v[5], 2)
	v3 := Utils.RoundTo(v[6], 2)

	if v1 != 1.02 || v2 != 2.06 || v3 != 4.33 {
		t.Errorf("Error with PValues: got %f, expected 1.02; got %f, expected 2.06; got %f, expected 4.33", v1, v2, v3)
	}
}

func TestParcel_crop(t *testing.T) {
	var mValues []fileio.MonthlyValues
	for i := 0; i < 12; i++ {
		mValues = append(mValues, fileio.MonthlyValues{Et: 0.1, Eff_precip: 1.1, Nir: 0.2, Dp: 0.1, Ro: 0.3, Precip: 0.9})
	}

	var mValues2 []fileio.MonthlyValues
	for i := 0; i < 12; i++ {
		mValues2 = append(mValues, fileio.MonthlyValues{Et: 0.2, Eff_precip: 1.2, Nir: 0.3, Dp: 0.2, Ro: 0.4, Precip: 1.0})
	}

	var aData []fileio.StationResults
	aData = append(aData, fileio.StationResults{Station: "SCTB", Soil: 622, Yr: 2014, Crop: 1, Tillage: 1, Irrigation: 1, MonthlyData: mValues})
	aData = append(aData, fileio.StationResults{Station: "SCTB", Soil: 622, Yr: 2014, Crop: 2, Tillage: 1, Irrigation: 1, MonthlyData: mValues2})

	nir, _, _, et := parcels.Crop(1, aData)

	if nir[1] != 0.2 {
		t.Errorf("Nir Values not being calclated correctly; got %f, expected 0.2", nir[1])
	}

	if et[1] != 0.1 {
		t.Errorf("Nir Values not being calclated correctly; got %f, expected 0.1", et[1])
	}
}
