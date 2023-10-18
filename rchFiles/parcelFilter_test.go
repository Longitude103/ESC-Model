package rchFiles

import (
	"bitbucket.org/longitude-103/wwumm-esc/parcels"
	"testing"
)

var p1 = parcels.Parcel{ParcelNo: 1, Yr: 2020, Nrd: "SP"}
var p2 = parcels.Parcel{ParcelNo: 2, Yr: 2020, Nrd: "SP"}
var p3 = parcels.Parcel{ParcelNo: 3, Yr: 2020, Nrd: "NP"}
var p4 = parcels.Parcel{ParcelNo: 4, Yr: 2021, Nrd: "NP"}
var p5 = parcels.Parcel{ParcelNo: 5, Yr: 2021, Nrd: "SP"}

var testParcelSlice = []parcels.Parcel{p1, p2, p3, p4, p5}

func Test_parcelFilterByYear(t *testing.T) {
	p, err := parcelFilterByYear(testParcelSlice, 2020)
	if len(p) != 3 || err != nil {
		t.Errorf("should have been 3 records but found %d, and error was nil but found %s", len(p), err)
	}

	p, err = parcelFilterByYear(testParcelSlice, 2019)
	if len(p) != 0 || err == nil {
		t.Errorf("should have returned no records and and error")
	}
}

func Test_parcelFilterById(t *testing.T) {
	p, err := parcelFilterById(testParcelSlice, 2, "SP")
	if p.ParcelNo != 2 && err != nil {
		t.Error("function should have return parcel 2 and no error")
	}

	_, err = parcelFilterById(testParcelSlice, 6, "NP")
	if err == nil {
		t.Error("function should return an error but didn't")
	}
}
