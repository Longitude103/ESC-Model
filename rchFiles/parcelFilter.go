package rchFiles

import (
	"github.com/Longitude103/ESC-Model/parcels"
	"errors"
)

// parcelFilterByYear is a function to filter our a parcels by a year and return a slice of those that are all of that year
func parcelFilterByYear(parcels []parcels.Parcel, yr int) (p []parcels.Parcel, err error) {
	if len(parcels) < 1 {
		return nil, errors.New("no parcels in slice")
	}

	for i := 0; i < len(parcels); i++ {
		if parcels[i].Yr == yr {
			p = append(p, parcels[i])
		}
	}

	if len(p) == 0 {
		return p, errors.New("no parcels found for that year")
	}

	return
}

// parcelFilterById is a function to filter the slice of parcels by an ID and NRD and return a single parcel.
func parcelFilterById(p []parcels.Parcel, id int, nrd string) (parcels.Parcel, error) {
	if len(p) < 1 {
		return parcels.Parcel{}, errors.New("no parcels in slice for id")
	}

	for i := 0; i < len(p); i++ {
		if p[i].ParcelNo == id && p[i].Nrd == nrd {
			return p[i], nil
		}
	}

	errMessage := "no parcel with that id"
	return parcels.Parcel{}, errors.New(errMessage)
}
