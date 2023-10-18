package parcels

import (
	"bitbucket.org/longitude-103/wwumm-esc/database"
	"database/sql"
	"fmt"
	"testing"
)

var dryParcelTest = Parcel{ParcelNo: 123, CertNum: sql.NullString{String: "1234", Valid: true}, Area: 40,
	Crop1: sql.NullInt64{Int64: 8, Valid: true}, Crop1Cov: sql.NullFloat64{Float64: 1.0, Valid: true}, CoeffZone: 4,
	DryEt: [12]float64{0, 0, 0, 1.1, 2.3, 2.8, 3.5, 4.1, 3.0, 1.1, 0, 0},
	Ro:    [12]float64{0, 0, 0, 0, .2, .65, .1, 0, 0, 0, 0, 0}, Dp: [12]float64{0, 0, 0, 0, .7, .95, 1.1, 0, 0, 0, 0, 0}}

var cCrops1 = database.CoeffCrop{Zone: 4, Crop: 8, DryEtAdj: .9, DryEtToRo: .5, DpAdj: 1, RoAdj: 1}

var TestCCrops = []database.CoeffCrop{cCrops1}

func Test_dryWaterBalanceWSPP(t *testing.T) {
	var RoSum, DpSum, EtSum float64
	for i := 0; i < 12; i++ {
		RoSum += dryParcelTest.Ro[i]
		DpSum += dryParcelTest.Dp[i]
		EtSum += dryParcelTest.DryEt[i]
	}

	transToRf := EtSum * (1 - TestCCrops[0].DryEtAdj)

	fmt.Println("Before WSPP")
	fmt.Printf("Total RO: %f, Total DP: %f\n", RoSum, DpSum)
	fmt.Printf("Total ET: %f, DryEtAdj: %f, ET Moved to RF: %f\n", EtSum, TestCCrops[0].DryEtAdj, transToRf)
	fmt.Printf("Trans to RO: %f, Trans to DP: %f\n", transToRf*TestCCrops[0].DryEtToRo, transToRf*(1-TestCCrops[0].DryEtToRo))
	fmt.Println("-------------------------------------------------")
	err := dryParcelTest.dryWaterBalanceWSPP(TestCCrops)
	if err != nil {
		t.Errorf("Error in method: %s", err)
	}

	RoSum = 0
	DpSum = 0
	for i := 0; i < 12; i++ {
		RoSum += dryParcelTest.Ro[i]
		DpSum += dryParcelTest.Dp[i]
	}

	fmt.Println("After WSPP")
	fmt.Printf("Total RO: %f, Total DP: %f\n", RoSum, DpSum)
	fmt.Printf("DryPRo: %v\nDPDp: %v\n", dryParcelTest.Ro, dryParcelTest.Dp)
	fmt.Println("-------------------------------------------------")
}
