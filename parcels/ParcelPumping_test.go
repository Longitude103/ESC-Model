package parcels_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Longitude103/ESC-Model/parcels"

	"github.com/Longitude103/ESC-Model/database"
	"github.com/Longitude103/ESC-Model/fileio"
	"github.com/joho/godotenv"
)

var u1 = parcels.Usage{Yr: 2014, Nrd: "np", CertNum: "3456", UseAF: 100.0}
var u2 = parcels.Usage{Yr: 2014, Nrd: "np", CertNum: "3459", UseAF: 240.0}
var u3 = parcels.Usage{Yr: 2014, Nrd: "np", CertNum: "3457", UseAF: 100.0}
var u4 = parcels.Usage{Yr: 2014, Nrd: "np", CertNum: "3458", UseAF: 200.0}

var testUsageSlice = []parcels.Usage{u1, u2, u3, u4}

func dbConnection(t testing.TB) *database.Setup {
	t.Helper()
	var myEnv map[string]string
	myEnv, err := godotenv.Read("../.env")
	if err != nil {
		t.Skipf("skipping database test: cannot load .env: %v", err)
	}

	var v *database.Setup
	v, err = database.NewSetup(myEnv, database.WithNoSQLite(), database.WithDebug())
	if err != nil {
		fmt.Printf("Error in NewSetup: %s", err)
	}

	if err = v.SetYears(1997, 1997); err != nil {
		fmt.Println("error setting years")
	}

	return v
}

func Test_distUsage(t *testing.T) {
	for i := 0; i < 12; i++ {
		testParcelSlice[0].Pump[i] = 0
	}

	var myEnv map[string]string
	myEnv, err := godotenv.Read("../.env")
	if err != nil {
		t.Skipf("skipping database test: cannot load .env: %v", err)
	}

	v, err := database.NewSetup(myEnv, database.WithNoSQLite(), database.WithDebug())
	if err != nil {
		fmt.Printf("Error in NewSetup: %s", err)
	}

	err = parcels.DistUsage(testUsageSlice, &testParcelSlice, v)
	if err != nil {
		t.Error("Function returned an error:", err)
	}

	total := 0.0
	for _, f := range testParcelSlice[0].Pump {
		total += f
	}

	if total < 9.89 || total > 9.9 {
		t.Errorf("Total pumping should have been 9.897 but got %f", total)
	}
}

func Test_ParcelPump(t *testing.T) {
	v := dbConnection(t)
	if err := v.SetYears(2010, 2010); err != nil {
		t.Error("Error Setting Years")
	}

	csResults, _ := fileio.LoadTextFiles("../testData/CropSimOutput", v.Logger)
	wStations, _ := database.GetWeatherStations(v.PgDb)
	//cCoefficients, _ := database.GetCoeffCrops(v)

	irrParcels, err := parcels.ParcelPump(v, csResults, wStations)
	if err != nil {
		v.Logger.Errorf("Error in Parcel Pumping: %s", err)
	}

	i := 0
	for _, parcel := range irrParcels {
		// was 310
		if parcel.ParcelNo == 4274 {
			i += 1
			v.Logger.Debug(parcel.String())
			v.Logger.Debug(parcel.NIRString())
			v.Logger.Debug(strings.Repeat("-", 100))
			v.Logger.Debug(parcel.SWString())
			v.Logger.Debug(parcel.PumpString())
			v.Logger.Debug(strings.Repeat("-", 100))
			v.Logger.Debug(parcel.RoString())
			v.Logger.Debug(parcel.DpString())
			v.Logger.Debug(strings.Repeat("=", 100))

			if i > 15 {
				break
			}
		}
	}

}

func Test_ParcelPumpPost97(t *testing.T) {
	v := dbConnection(t)
	if err := v.SetYears(2017, 2017); err != nil {
		t.Error("Error Setting Years")
	}
	v.Post97 = true

	csResults, _ := fileio.LoadTextFiles("../testData/CropSimOutput", v.Logger)
	wStations, _ := database.GetWeatherStations(v.PgDb)
	//cCoefficients, _ := database.GetCoeffCrops(v)

	irrParcels, err := parcels.ParcelPump(v, csResults, wStations)
	if err != nil {
		v.Logger.Errorf("Error in Parcel Pumping: %s", err)
	}

	//i := 0
	for _, parcel := range irrParcels {
		// was 310
		//if parcel.ParcelNo == 1022 {
		//	i += 1
		//	v.Logger.Debug(parcel.String())
		//	v.Logger.Debug(parcel.NIRString())
		//	v.Logger.Debug(strings.Repeat("-", 100))
		//	v.Logger.Debug(parcel.SWString())
		//	v.Logger.Debug(parcel.PumpString())
		//	v.Logger.Debug(strings.Repeat("-", 100))
		//	v.Logger.Debug(parcel.RoString())
		//	v.Logger.Debug(parcel.DpString())
		//	v.Logger.Debug(strings.Repeat("=", 100))
		//
		//	if i > 15 {
		//		break
		//	}
		//}

		if parcel.IsComingled() {
			fmt.Printf("%+v\n\n", parcel)
		}
	}

}
