package rchFiles

import (
	"errors"
	"github.com/Longitude103/ESC-Model/parcels"
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
