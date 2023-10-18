package parcels_test

import (
	"bitbucket.org/longitude-103/wwumm-esc/database"
	"bitbucket.org/longitude-103/wwumm-esc/parcels"
	"testing"
)

var (
	eff1         = database.Efficiency{Yr: 2014, AeFlood: 0.65, AeSprinkler: 0.85}
	eff2         = database.Efficiency{Yr: 2013, AeFlood: 0.60, AeSprinkler: 0.80}
	efficiencies = []database.Efficiency{eff1, eff2}
)

func Test_filterEff(t *testing.T) {
	f, s := parcels.FilterEff(efficiencies, 2014)

	if f != 0.65 || s != 0.85 {
		t.Errorf("filter Efficiencies not correct for 2014: got %f, expected 0.65; got %f, expected 0.85", f, s)
	}
}

func TestParcel_setAppEfficiency(t *testing.T) {
	p1.SetAppEfficiency(efficiencies, 2013)
	p3.SetAppEfficiency(efficiencies, 2013)

	if p1.AppEff != 0.8 {
		t.Errorf("AppEff not set correctly: got %f, expected 0.8", p1.AppEff)
	}

	if p3.AppEff != 0.6 {
		t.Errorf("AppEff not set correctly: got %f, expected 0.8", p3.AppEff)
	}
}
