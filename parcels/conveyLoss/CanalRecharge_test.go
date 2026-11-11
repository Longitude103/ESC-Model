package conveyLoss

import (
	"testing"
)

func TestCanalRecharge(t *testing.T) {
	v := dbConnection(t)
	v.AppDebug = true

	if err := v.SetYears(2011, 2011); err != nil {
		t.Error("Failed to set years")
	}

	if err := CanalRecharge(v); err != nil {
		t.Error("Canal Recharge Failed: ", err)
	}

	// Results should be...
	// CellDiversions: [{Diversion:{CanalId:11 DivDate:{Time:2011-04-01 00:00:00 +0000 UTC Valid:true} DivAmount:{Float64:7343 Valid:true}} LossPercent:{Float64:0.49 Valid:true}} ...]
	// Div: {Diversion:{CanalId:11 DivDate:{Time:2011-04-01 00:00:00 +0000 UTC Valid:true} DivAmount:{Float64:7343 Valid:true}} LossPercent:{Float64:0.49 Valid:true}}
	// Structure Loss Percent: 0.490000, StructureLoss: 3598.070000, Factor: 0.001515
	// Cell Data: {CanalId:11 CanalType:Canal DistId:25 Eff:{Float64:0.51 Valid:true} Node:12987 CellArea:10 StLength:643.4954683102558 CFlag:1 DnrFact:{Float64:0.00301318181818 Valid:true} SatFact:{Float64:0 Valid:false} UsgsFact:{Float64:0 Valid:false} CLinkId:11 CanalEff:{Float64:0.51 Valid:true} LatCount:{Int64:43 Valid:true} TotalLatLn:{Float64:619549.6820054883 Valid:true} TotalCanLn:424794.39885072457}
	// Cell Result: {Node:12987 Size:10 Dt:2011-04-01 00:00:00 +0000 UTC FileType:124 Result:10.811066137045628}
	// Check Total is 7136.771845 .... which when divided by 1.9835 = 3598.07 which equals structure loss
}

func Test_getLossPercent(t *testing.T) {
	v := dbConnection(t)
	if err := v.SetYears(2016, 2016); err != nil {
		t.Error("Failed to set years")
	}

	_, div, err := getDiversions(v)
	if err != nil {
		t.Error("Error in getDiversions: ", err)
	}

	ef := getLossPercent(div, v)
	for _, e := range ef {
		switch e.CanalId {
		case 29:
			if e.LossPercent.Float64 != 0.24 {
				t.Errorf("canal %d should have loss percent of 0.24, but got %f", e.CanalId, e.LossPercent.Float64)
			}
		case 39:
			if e.LossPercent.Float64 != 0.53 {
				t.Errorf("canal %d should have loss percent of 0.53, but got %f", e.CanalId, e.LossPercent.Float64)
			}
		case 11:
			if e.LossPercent.Float64 != 0.49 {
				t.Errorf("canal %d should have loss percent of 0.49, but got %f", e.CanalId, e.LossPercent.Float64)
			}
		}
	}
}
