package rchFiles

import (
	"github.com/Longitude103/ESC-Model/parcels"
)

// parcelKey identifies a parcel within a year. A parcel can have several rows (one per soil and coefficient zone
// combination it covers), so all rows share this key.
type parcelKey struct {
	id  int
	nrd string
}

// indexParcels groups a year's parcel rows by parcel id and nrd so the cell loops can look them up quickly.
func indexParcels(p []parcels.Parcel) map[parcelKey][]parcels.Parcel {
	idx := make(map[parcelKey][]parcels.Parcel, len(p))
	for _, row := range p {
		k := parcelKey{id: row.ParcelNo, nrd: row.Nrd}
		idx[k] = append(idx[k], row)
	}

	return idx
}

// cellParcelRow returns the values that belong to the part of a parcel that falls in a model cell.
//
// Parcel RO, DP and ET are acre-feet for one soil / coefficient zone row of the parcel (scaled by SoilArea), so a cell
// must use the row that matches its own soil and coefficient zone and divide by that row's SoilArea. If no row matches
// (or the row has no area), it falls back to the whole parcel: the sum of every row divided by the parcel area. Either
// way the parcel's full RO and DP is distributed across its cells. matched reports whether a soil row was found.
func cellParcelRow(rows []parcels.Parcel, soil, zone int) (row parcels.Parcel, area float64, matched bool) {
	for _, r := range rows {
		if r.SoilCode == soil && r.CoeffZone == zone && r.SoilArea > 0 {
			return r, r.SoilArea, true
		}
	}

	// fallback: combine all the rows of the parcel
	row = rows[0]
	row.Ro, row.Dp, row.Et = [12]float64{}, [12]float64{}, [12]float64{}
	for _, r := range rows {
		for m := 0; m < 12; m++ {
			row.Ro[m] += r.Ro[m]
			row.Dp[m] += r.Dp[m]
			row.Et[m] += r.Et[m]
		}
		area += r.SoilArea
	}

	if area <= 0 {
		area = rows[0].Area
	}

	return row, area, false
}
