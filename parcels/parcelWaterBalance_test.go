package parcels_test

import (
	"github.com/Longitude103/ESC-Model/Utils"
	"github.com/Longitude103/ESC-Model/parcels"
	"database/sql"
	"fmt"
	"strings"
	"testing"
)

var p10 = parcels.Parcel{ParcelNo: 1234, AppEff: 0.85,
	Nir:       [12]float64{0, 0, 0, 0, 0.2, 0.4, 0.8, 0.8, 0.5, 0, 0, 0},
	DryEt:     [12]float64{0, 0, 0, 1.1, 0.05, 0.1, 0.2, 0.2, 0.1, 0, 0, 0},
	Et:        [12]float64{0, 0, 0, 1.2, 1.2, 2.5, 4.5, 4.5, 3, 0, 0, 0},
	Pump:      [12]float64{0, 0, 0, 0, 0, 12.3, 21.4, 18.9, 0, 0, 0, 0},
	Ro:        [12]float64{0, 0, 0, 0, 0, 0, 0, 0.35, 0.87, 0, 0, 0},
	Dp:        [12]float64{0, 0, 0, 0, 0, 0, 1.5, 1.1, 1.3, 0, 0, 0},
	SWDel:     [12]float64{0, 0, 0, 0, 1, 13, 15, 14, 5, 0, 0, 0},
	CoeffZone: 2, SoilCode: 622, Area: 40.0, IrrType: sql.NullString{String: "SPRINKLER", Valid: true},
	Crop1:    sql.NullInt64{Int64: 1, Valid: true},
	Crop1Cov: sql.NullFloat64{Float64: 1, Valid: true}, Crop2: sql.NullInt64{Int64: 0, Valid: false},
	Crop2Cov: sql.NullFloat64{Float64: 0, Valid: false}, Crop3: sql.NullInt64{Int64: 0, Valid: false},
	Crop3Cov: sql.NullFloat64{Float64: 0, Valid: false}, Crop4: sql.NullInt64{Int64: 0, Valid: false},
	Crop4Cov: sql.NullFloat64{Float64: 0, Valid: false}, Nrd: "np", Yr: 2014,
	CertNum: sql.NullString{String: "3456", Valid: true}, PointX: 41.4, PointY: 103.0,
	Sw: sql.NullBool{Bool: true, Valid: true}, Gw: sql.NullBool{Bool: true, Valid: true}}

var p11 = parcels.Parcel{ParcelNo: 1234, AppEff: 0.85,
	Nir:       [12]float64{0, 0, 0, 0, 0.2, 0.4, 0.8, 0.8, 0.5, 0, 0, 0},
	DryEt:     [12]float64{0, 0, 0, 0, 0.05, 0.1, 0.2, 0.2, 0.1, 0, 0, 0},
	Et:        [12]float64{0, 0, 0, 0, 1.2, 2.5, 4.5, 4.5, 3, 0, 0, 0},
	SWDel:     [12]float64{0, 0, 0, 0, 0, 12, 15, 14, 5, 0, 0, 0},
	Ro:        [12]float64{0, 0, 0, 0, 0, 0, 0, 0.35, 0.87, 0, 0, 0},
	Dp:        [12]float64{0, 0, 0, 0, 0, 0, 1.5, 1.1, 1.3, 0, 0, 0},
	CoeffZone: 3, Crop1: sql.NullInt64{Int64: 8, Valid: true},
	Crop1Cov: sql.NullFloat64{Float64: 0.5, Valid: true}, Crop2: sql.NullInt64{Int64: 5, Valid: true},
	Crop2Cov: sql.NullFloat64{Float64: 0.5, Valid: true}, Crop3: sql.NullInt64{Int64: 0, Valid: false},
	Crop3Cov: sql.NullFloat64{Float64: 0, Valid: false}, Crop4: sql.NullInt64{Int64: 0, Valid: false},
	Crop4Cov: sql.NullFloat64{Float64: 0, Valid: false}, Nrd: "np", Yr: 2014,
	CertNum: sql.NullString{String: "3456", Valid: true}, PointX: 40.21, PointY: 103.0,
	Sw: sql.NullBool{Bool: true, Valid: true}, Gw: sql.NullBool{Bool: false, Valid: true}}

// p12 is the groundwater only cell made into a parcel from the TFG Example document
var p12 = parcels.Parcel{ParcelNo: 159988, AppEff: 0.65,
	Nir:       [12]float64{0, 0, 0, 0, 0, 0, 4.98, 4.31, 1.65, 0, 0, 0},
	DryEt:     [12]float64{0.24, 0.62, 0.39, 1.36, 1.82, 5.13, 4.55, 2.66, 1.16, 0.70, 0.66, 0.19},
	Et:        [12]float64{0.27, 0.33, 0.82, 1.36, 1.82, 5.13, 7.77, 7.21, 4.02, 0.44, 0.51, 0.23},
	Pump:      [12]float64{0, 0, 0, 0, 0, 0, 2.34, 2.32, 1.4, 0, 0, 0},
	Ro:        [12]float64{0, 0, 0, 1.04, 0.73, 1.81, 0, 0.11, 0, 0.03, 0, 0},
	Dp:        [12]float64{0, 0, 0, 0, 0, 0.39, 0, 0, 0, 0.01, 0, 0},
	CoeffZone: 2, SoilCode: 622, Area: 40.0, IrrType: sql.NullString{String: "FLOOD", Valid: true},
	Crop1: sql.NullInt64{Int64: 1, Valid: true}, Crop1Cov: sql.NullFloat64{Float64: 1, Valid: true},
	Crop2:    sql.NullInt64{Int64: 0, Valid: false},
	Crop2Cov: sql.NullFloat64{Float64: 0, Valid: false}, Crop3: sql.NullInt64{Int64: 0, Valid: false},
	Crop3Cov: sql.NullFloat64{Float64: 0, Valid: false}, Crop4: sql.NullInt64{Int64: 0, Valid: false},
	Crop4Cov: sql.NullFloat64{Float64: 0, Valid: false}, Nrd: "np", Yr: 2014,
	CertNum: sql.NullString{String: "3456", Valid: true}, PointX: 41.4, PointY: 102.5,
	Sw: sql.NullBool{Bool: false, Valid: true}, Gw: sql.NullBool{Bool: true, Valid: true}}

func TestParcel_WaterBalanceWWSP(t *testing.T) {
	v := dbConnection(t)

	fmt.Println(strings.Repeat("=", 120))
	fmt.Println("RO is:", p10.Ro)
	fmt.Println("Dp is:", p10.Dp)

	err := p10.WaterBalanceWSPP(v)
	if err != nil {
		t.Errorf("Error in WSPP Method of %s", err)
	}

	fmt.Println("Post Process RO is:", p10.Ro)
	fmt.Println("Post Process Dp is:", p10.Dp)
	fmt.Println(strings.Repeat("=", 120))

}

func TestParcel_WaterBalanceWWSP_SWOnly(t *testing.T) {
	v := dbConnection(t)
	fmt.Println(strings.Repeat("=", 120))
	fmt.Println("RO is:", p11.Ro)
	fmt.Println("Dp is:", p11.Dp)

	err := p11.WaterBalanceWSPP(v)
	if err != nil {
		t.Errorf("Error in WSPP Method of %s", err)
	}

	fmt.Println("Post Process RO is:", p11.Ro)
	fmt.Println("Post Process Dp is:", p11.Dp)
	fmt.Println(strings.Repeat("=", 120))
}

func TestParcel_WaterBalanceWWSP_GWOnly(t *testing.T) {
	v := dbConnection(t)
	fmt.Println(strings.Repeat("=", 120))
	fmt.Println("RO is:", p12.Ro)
	fmt.Println("Dp is:", p12.Dp)

	err := p12.WaterBalanceWSPP(v)
	if err != nil {
		t.Errorf("Error in WSPP Method of %s", err)
	}

	fmt.Println("Post Process RO is:", p12.Ro)
	fmt.Println("Post Process Dp is:", p12.Dp)

	// need to test July and August
	if Utils.RoundTo(p12.Ro[6], 2) != 0.66 || Utils.RoundTo(p12.Ro[7], 2) != 0.70 {
		t.Errorf("July RO got %g, expected 0.66", Utils.RoundTo(p12.Ro[6], 2))
		t.Errorf("August RO got %g, expected 0.70", Utils.RoundTo(p12.Ro[7], 2))
	}

	if Utils.RoundTo(p12.Dp[6], 2) != 0.66 || Utils.RoundTo(p12.Dp[7], 2) != 0.15 {
		t.Errorf("July DP got %g, expected 0.66", Utils.RoundTo(p12.Dp[6], 2))
		t.Errorf("August DP got %g, expected 0.15", Utils.RoundTo(p12.Dp[7], 2))
	}
	fmt.Println(strings.Repeat("=", 120))
}

func TestParcel_setGirFact(t *testing.T) {
	eff := 0.85

	gir, fsl := parcels.SetGirFact(eff)
	if gir != 1/0.95 || fsl != 0.02 {
		t.Errorf("Error in setGirFact eff of %g, gir got: %g, should be 1.05263..; fsl got: %g, should be 0.02", eff, gir, fsl)
	}

	eff = 0.65
	gir, fsl = parcels.SetGirFact(eff)

	if gir != 1/0.75 || fsl != 0.05 {
		t.Errorf("Error in setGirFact with eff of %g, gir got: %g, should be 1.3333..; fsl got: %g, should be 0.05", eff, gir, fsl)
	}
}

func TestParcel_sumAnnual(t *testing.T) {
	data := [12]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

	result := parcels.SumAnnual(data)

	if result != 78 {
		t.Errorf("should have gotten 78 but got %g instead", result)
	}
}

func TestParcel_setAppWat(t *testing.T) {
	sw := [12]float64{50, 0, 0, 0, 0, 100, 0, 0, 0, 0, 0, 0}
	gw := [12]float64{50, 0, 0, 0, 100, 0, 0, 0, 0, 0, 0, 0}
	fsl := 0.02

	appWat, sL, pslIrr := parcels.SetAppWat(sw, gw, fsl)

	if appWat[0] != 100 || appWat[4] != 100 || appWat[5] != 100 {
		t.Errorf("All appWat values should be 100 but got 0: %g, 4:%g, 5:%g", appWat[0], appWat[4], appWat[5])
	}

	if sL[0] != 2 || sL[4] != 2 || sL[5] != 2 {
		t.Errorf("All sL values should be 2 but got 0: %g, 4:%g, 5:%g", sL[0], sL[4], sL[5])
	}

	if pslIrr[0] != 98 || pslIrr[4] != 98 || pslIrr[5] != 98 {
		t.Errorf("All pslIrr values should be 98 but got 0: %g, 4:%g, 5:%g", pslIrr[0], pslIrr[4], pslIrr[5])
	}
}

func TestParcel_setRoDpWt(t *testing.T) {
	ro := [12]float64{10, 75, 5, 100, 0, 0, 0, 0, 0, 0, 0, 0}
	dp := [12]float64{30, 25, 100, 5, 0, 0, 0, 0, 0, 0, 0, 0}

	result, err := parcels.SetRoDpWt(ro, dp)

	if err != nil {
		t.Error(err)
	}

	if result[0] != 0.25 || result[1] != 0.75 || result[2] != 0.2 || result[3] != 0.8 || result[4] != 0.5 {
		t.Errorf("wieights shoudl be 0.25, 0.75, 0.2, 0.8, 0.5... and got %v", result)
	}
}

func TestParcel_setInitialRoDp(t *testing.T) {
	ro, dp := parcels.SetInitialRoDp(p12.Ro, p12.Dp, 1, 1)

	if Utils.RoundTo(ro[0], 3) != 0.0 || Utils.RoundTo(ro[1], 3) != 0.0 || Utils.RoundTo(ro[2], 3) != 0.0 || Utils.RoundTo(ro[3], 3) != 1.040 {
		t.Errorf("incorrect initial values for RO, should be 0.0, 0.0, 0.0, 1.040... and got %v", ro)
	}

	if Utils.RoundTo(dp[0], 3) != 0.0 || Utils.RoundTo(dp[1], 3) != 0.0 || Utils.RoundTo(dp[2], 3) != 0.0 || Utils.RoundTo(dp[5], 3) != 0.390 {
		t.Errorf("incorrect initial values for DP, should be 0.0, 0.0, 0.0, 0.390... and got %v", dp)
	}
}

func TestParcel_setPreGain(t *testing.T) {
	et := [12]float64{12, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	dryEt := [12]float64{6, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	appWat := [12]float64{3, 3, 3, 3, 0, 0, 0, 0, 0, 0, 0, 0}
	psl := [12]float64{2, 2, 2, 2, 0, 0, 0, 0, 0, 0, 0, 0}

	gainApWat, gainPsl, gainIrrEt, gainDryEt := parcels.SetPreGain(et, dryEt, appWat, psl)

	if gainApWat != 3 || gainPsl != 2 || gainIrrEt != 12 || gainDryEt != 6 {
		t.Errorf("error in setPreGain, gainApWat: %g, expecting 3; gainPsl: %g, expecting 2; gainIrrEt: %g, "+
			"expecting 12; gainDryEt: %g, expecting 6", gainApWat, gainPsl, gainIrrEt, gainDryEt)
	}
}

func TestParcel_setEtGain(t *testing.T) {
	appWat := 6.07
	psl := 5.76
	cir := 10.22
	gir := 14.59
	eff := 0.65
	irrEt := 19.00
	dryEt := 8.78

	gain, err := parcels.SetEtGain(cir, psl, gir, appWat, eff, irrEt, dryEt)
	if err != nil {
		t.Error(err)
	}

	if Utils.RoundTo(gain, 2) != 3.95 {
		t.Errorf("error in gain function, got: %g, expected 3.95", Utils.RoundTo(gain, 2))
	}
}

func TestParcel_distEtGain(t *testing.T) {
	etDry := [12]float64{0.24, 0.62, 0.39, 1.36, 1.82, 5.13, 4.55, 2.66, 1.16, 0.7, 0.66, 0.19}
	etIrr := [12]float64{0.27, 0.33, 0.82, 1.36, 1.82, 5.13, 7.77, 7.21, 4.02, 0.44, 0.51, 0.23}
	psl := [12]float64{0, 0, 0, 0, 0, 0, 2.22, 2.21, 1.33, 0, 0, 0}
	gain := 3.94

	dist, err := parcels.DistEtGain(gain, psl, etIrr, etDry)
	if err != nil {
		t.Error(err)
	}

	if Utils.RoundTo(dist[6], 3) != 1.193 || Utils.RoundTo(dist[7], 3) != 1.686 {
		t.Errorf("distETGain calculated incorrectly: July %g, expected 1.193; August %g, expected 1.686", Utils.RoundTo(dist[6], 3), Utils.RoundTo(dist[7], 3))
	}

	psl = [12]float64{0, 0, 0, 0, 0.98, 24.794, 35.672, 32.242, 4.9, 0, 0, 0}
	gain = 15.05

	dist, err = parcels.DistEtGain(gain, psl, p10.Et, p10.DryEt)
	if err != nil {
		t.Error(err)
	}

	if Utils.RoundTo(dist[3], 3) != 0.170 {
		t.Errorf("distETGain calculated incorrectly when remaining gain present: April %g, expected 0.170", Utils.RoundTo(dist[3], 3))
	}
}

func TestParcel_setEtBase(t *testing.T) {
	etDry := [12]float64{0.24, 0.62, 0.39, 1.36, 1.82, 5.13, 4.55, 2.66, 1.16, 0.7, 0.66, 0.19}
	etIrr := [12]float64{0.27, 0.33, 0.82, 1.36, 1.82, 5.13, 7.77, 7.21, 4.02, 0.44, 0.51, 0.23}
	psl := [12]float64{0, 0, 0, 0, 0, 0, 2.22, 2.21, 1.33, 0, 0, 0}

	base := parcels.SetEtBase(psl, etIrr, etDry)

	if base[5] != 5.13 || base[6] != 4.55 {
		t.Errorf("base calculated incorrect: June base: %g, expecting 5.13; July base: %g, expecting 4.55", base[5], base[6])
	}
}

func TestParcel_setET(t *testing.T) {
	etBase := [12]float64{0.27, 0.33, 0.82, 1.36, 1.82, 5.13, 4.55, 2.66, 1.16, 0.44, 0.51, 0.23}
	etGain := [12]float64{0, 0, 0, 0, 0, 0, 1.19, 1.69, 1.06, 0, 0, 0}

	et := parcels.SetET(etBase, etGain)

	if et[5] != 5.13 || et[6] != 5.74 {
		t.Errorf("et calculated incorrect: June ET: %g, expecting 5.13; July ET: %g, expecting 5.74", et[5], et[6])
	}
}

func TestParcel_setDeltaET(t *testing.T) {
	etIrr := [12]float64{0.26, 0.31, 0.78, 1.29, 1.73, 4.87, 5.46, 4.13, 2.11, 0.42, 0.48, 0.22}
	factor := 0.95

	delta := parcels.SetDeltaET(etIrr, factor)

	if Utils.RoundTo(delta[0], 3) != 0.013 || Utils.RoundTo(delta[5], 3) != 0.244 {
		t.Errorf("setDeltaET calculated incorrect: Jan delta: %g, expected 0.013; June delat: %g, expected 0.244", Utils.RoundTo(delta[0], 3), Utils.RoundTo(delta[5], 3))
	}
}

func TestParcel_distDeltaET(t *testing.T) {
	deltaET := [12]float64{0.01, 0.02, 0.04, 0.07, 0.09, 0.26, 0.29, 0.22, 0.11, 0.02, 0.03, 0.01}
	roDpWt := [12]float64{0.5, 0.5, 0.5, 0.8, 0.8, 0.8, 0.5, 0.8, 0.5, 0.8, 0.5, 0.5}

	ro, dp := parcels.DistDeltaET(deltaET, roDpWt)

	if Utils.RoundTo(ro[3], 3) != 0.056 || Utils.RoundTo(ro[5], 3) != 0.208 {
		t.Errorf("distDeltaET calculated RO incorrect: April: %g, expected 0.056; Jun: %g, expected 0.208", Utils.RoundTo(ro[3], 3), Utils.RoundTo(ro[5], 3))
	}

	if Utils.RoundTo(dp[3], 3) != 0.014 || Utils.RoundTo(dp[5], 3) != 0.052 {
		t.Errorf("distDeltaET calculated DP incorrect: April: %g, expected 0.014; Jun: %g, expected 0.052", Utils.RoundTo(dp[3], 3), Utils.RoundTo(dp[5], 3))
	}
}

func TestParcel_excessIrrReturnFlow(t *testing.T) {
	roDpWt := [12]float64{0.5, 0.5, 0.5, 0.8, 0.8, 0.8, 0.5, 0.8, 0.5, 0.8, 0.5, 0.5}
	etGain := [12]float64{0, 0, 0, 0, 0, 0, 1.19, 1.69, 1.06, 0, 0, 0}
	psl := [12]float64{0, 0, 0, 0, 0, 0, 2.22, 2.21, 1.33, 0, 0, 0}

	ro, dp := parcels.ExcessIrrReturnFlow(psl, etGain, roDpWt)

	if Utils.RoundTo(ro[6], 3) != 0.515 || Utils.RoundTo(ro[7], 3) != 0.416 {
		t.Errorf("excessIrrReturnFlow RO calculated incorrect: July %g, expected 0.515; Aug %g, expected 0.416", Utils.RoundTo(ro[6], 3), Utils.RoundTo(ro[7], 3))
	}

	if Utils.RoundTo(dp[6], 3) != 0.515 || Utils.RoundTo(dp[7], 3) != 0.104 {
		t.Errorf("excessIrrReturnFlow DP calculated incorrect: July %g, expected 0.515; Aug %g, expected 0.104", Utils.RoundTo(dp[6], 3), Utils.RoundTo(dp[7], 3))
	}

}

func TestParcel_sumReturnFlows(t *testing.T) {
	v1 := [12]float64{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	v2 := [12]float64{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	v3 := [12]float64{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	v4 := parcels.SumReturnFlows(v1, v2, v3)

	if v4[0] != 3 || v4[1] != 6 || v4[2] != 9 || v4[3] != 0 {
		t.Errorf("sumReturnFlows not correct: got %v, expected 3, 6, 9, 0, 0, ...", v4)
	}
}
