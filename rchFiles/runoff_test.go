package rchFiles

import (
	"math"
	"testing"
)

// Runoff that infiltrates on its way to the stream (RO x loss factor x percent to recharge) must reach recharge; it
// was always zero before.
func TestCalcRo_RunoffReachesRecharge(t *testing.T) {
	// 2 in RO on 120 acres with full station weight = 20 af; aRo 1, loss factor 0.5, 30% to recharge
	runOff, roToRch := calcRo(2, 0, 1, 120, 1, 0.5, 0.3)
	if math.Abs(runOff-10) > 1e-9 || math.Abs(roToRch-3) > 1e-9 {
		t.Errorf("calcRo got runOff %f roToRch %f, want 10 and 3", runOff, roToRch)
	}

	runOff, roToRch = calcRoDryland(20, 0, 1, 0.5, 0.3)
	if math.Abs(runOff-10) > 1e-9 || math.Abs(roToRch-3) > 1e-9 {
		t.Errorf("calcRoDryland got runOff %f roToRch %f, want 10 and 3", runOff, roToRch)
	}
}
