package parcels_test

import (
	"database/sql"
	"github.com/Longitude103/ESC-Model/Utils"
	"github.com/Longitude103/ESC-Model/database"
	"github.com/Longitude103/ESC-Model/parcels"
	"testing"
)

var (
	cCrop1 = database.CoeffCrop{
		Zone:         1,
		Crop:         1,
		DryEtAdj:     0.1,
		IrrEtAdj:     0.2,
		NirAdjFactor: 0.3,
		FslGW:        0.4,
		DryEtToRo:    0.5,
		FslSW:        0.6,
		PerToRch:     0.7,
		DpAdj:        0.8,
		RoAdj:        0.9,
	}
	cCrop2 = database.CoeffCrop{
		Zone:         2,
		Crop:         2,
		DryEtAdj:     1.1,
		IrrEtAdj:     1.2,
		NirAdjFactor: 1.3,
		FslGW:        1.4,
		DryEtToRo:    1.5,
		FslSW:        1.6,
		PerToRch:     1.7,
		DpAdj:        1.8,
		RoAdj:        1.9,
	}
	cCrop3 = database.CoeffCrop{
		Zone:         2,
		Crop:         1,
		DryEtAdj:     2.1,
		IrrEtAdj:     2.2,
		NirAdjFactor: 0.95,
		FslGW:        2.4,
		DryEtToRo:    2.5,
		FslSW:        2.6,
		PerToRch:     2.7,
		DpAdj:        2.8,
		RoAdj:        2.9,
	}

	cCrops = []database.CoeffCrop{cCrop1, cCrop2, cCrop3}
)

func TestPumping_adjFactor(t *testing.T) {
	result, err := parcels.AdjFactor(cCrops, 1, 1, database.DryET)
	if result != 0.1 {
		t.Errorf("adjFactor error: got DryETAdj %f, expected 0.1, error: %s", result, err)
	}

	result, err = parcels.AdjFactor(cCrops, 2, 2, database.NirEt)
	if result != 1.3 {
		t.Errorf("adjFactor error: got NirETAdj %f, expected 1.3, error: %s", result, err)
	}
}

func TestPumping_adjustmentFactor(t *testing.T) {
	result, err := parcels.AdjustmentFactor(&p3, cCrops, database.NirEt)

	if result != 0.95 {
		t.Errorf("adjustmentFactor error: got %f, expected 0.95; error: %s", result, err)
	}

	result, err = parcels.AdjustmentFactor(&p2, cCrops, database.NirEt)

	if result != 0 || err == nil {
		t.Errorf("adjustmentFactor should return a zero with error")
	}
}

func TestPumping_estimatePumping(t *testing.T) {
	// EstimatePumping only reads the run mode flags, so no database is needed. Copies keep the shared test parcels
	// (which other tests change) out of it.
	v := &database.Setup{}

	// commingled sprinkler parcel: (NIR - SW x eff) / eff
	sp := p1
	sp.AppEff = 0.85
	sp.Pump = [12]float64{}
	sp.Nir = [12]float64{0, 0, 0, 0, 3, 25, 40, 45, 20, 0, 0, 0}
	sp.Subarea = sql.NullString{String: "FA", Valid: true}

	if err := sp.EstimatePumping(v); err != nil {
		t.Errorf("Should not return error")
	}

	// May (3 - 1 x 0.85) / 0.85, June (25 - 13 x 0.85) / 0.85
	mayPump := Utils.RoundTo(sp.Pump[4], 2)
	junePump := Utils.RoundTo(sp.Pump[5], 2)
	if mayPump != 2.53 || junePump != 16.41 {
		t.Errorf("Pumping is not calculated correctly: May got %f, expected 2.53; June got %f, expected 16.41", mayPump, junePump)
	}

	// metered years (NP after 2016) keep their pumping
	gwo := p3
	gwo.AppEff = 0.65
	gwo.Yr = 2017
	_ = gwo.EstimatePumping(v)
	if got := Utils.RoundTo(gwo.Pump[6], 2); got != 2.34 {
		t.Errorf("Wanted Pumping 2.34 (shouldn't have changed), Got Pumping %f", got)
	}

	// post 97 runs estimate GW only parcels: the TFG example, July NIR 4.98 / 0.65 flood efficiency = 7.66
	v.Post97 = true
	_ = gwo.EstimatePumping(v)
	if got := Utils.RoundTo(gwo.Pump[6], 2); got != 7.66 {
		t.Errorf("Wanted Pumping 7.66, Got Pumping %f", got)
	}
}

func TestPumping_estimatePumpingNoEfficiency(t *testing.T) {
	p := parcels.Parcel{ParcelNo: 1, Nrd: "np", Yr: 2000, Gw: sql.NullBool{Valid: true, Bool: true},
		Nir: [12]float64{6: 1}}
	if err := p.EstimatePumping(&database.Setup{}); err == nil {
		t.Error("expected an error when the parcel has no application efficiency")
	}
}
