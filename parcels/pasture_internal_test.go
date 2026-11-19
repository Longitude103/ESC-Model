package parcels

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/Longitude103/ESC-Model/fileio"
)

func TestDryLandCropForIrrigatedPasture(t *testing.T) {
	dryData := []fileio.StationResults{
		{Crop: 13, Irrigation: int(DryLand), MonthlyData: []fileio.MonthlyValues{{Et: 1.5}}},
		{Crop: 8, Irrigation: int(DryLand), MonthlyData: []fileio.MonthlyValues{{Et: 0.7}}},
	}

	if _, _, _, et := Crop(dryLandCrop(12), dryData); et[0] != 1.5 {
		t.Errorf("irrigated pasture dry ET should come from dryland pasture: got %f, want 1.5", et[0])
	}
	if _, _, _, et := Crop(dryLandCrop(8), dryData); et[0] != 0.7 {
		t.Errorf("other crops keep their own dry ET: got %f, want 0.7", et[0])
	}
}

func TestCheckDryLandCrops(t *testing.T) {
	ok := Parcel{ParcelNo: 1, Crop1: sql.NullInt64{Int64: 13, Valid: true}, Crop2: sql.NullInt64{Int64: 15, Valid: true}}
	if err := ok.CheckDryLandCrops(); err != nil {
		t.Errorf("pasture and fallow are fine on dryland, got %s", err)
	}

	bad := Parcel{ParcelNo: 2, Crop1: sql.NullInt64{Int64: 8, Valid: true}, Crop3: sql.NullInt64{Int64: 12, Valid: true}}
	if err := bad.CheckDryLandCrops(); err == nil {
		t.Error("dryland parcel with irrigated pasture should be an error")
	}
}

func TestCheckIrrigatedCrops(t *testing.T) {
	ok := Parcel{ParcelNo: 1, Crop1: sql.NullInt64{Int64: 12, Valid: true}}
	if err := ok.CheckIrrigatedCrops(); err != nil {
		t.Errorf("irrigated pasture is fine on an irrigated parcel, got %s", err)
	}

	bad := Parcel{ParcelNo: 4417, Nrd: "np", Yr: 2005, Area: 132.4, CertNum: sql.NullString{String: "A-1234", Valid: true},
		Crop1: sql.NullInt64{Int64: 8, Valid: true}, Crop2: sql.NullInt64{Int64: 13, Valid: true}}
	err := bad.CheckIrrigatedCrops()
	if err == nil {
		t.Fatal("irrigated parcel with pasture should be an error")
	}
	for _, want := range []string{"Parcel 4417", "NRD NP", "year 2005", "cert A-1234", "crop2", "irrigated pasture (crop 12)"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should contain %q", err, want)
		}
	}

	all := checkCrops([]Parcel{bad, ok, bad}, (*Parcel).CheckIrrigatedCrops)
	if all == nil || strings.Count(all.Error(), "Parcel 4417") != 2 {
		t.Errorf("checkCrops should report every bad parcel, got %v", all)
	}
}
