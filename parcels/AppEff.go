package parcels

import (
	"github.com/Longitude103/ESC-Model/database"
	"strings"
)

// SetAppEfficiency is a Parcel method that uses the parcel IrrType and year to determine the application efficiency of the
// parcel and sets that in the Parcel struct. It takes a slice of Efficiency and year as they very through the study period.
func (p *Parcel) SetAppEfficiency(efficiencies []database.Efficiency, year int) {
	f, s := FilterEff(efficiencies, year)

	compareString := strings.ToLower(p.IrrType.String)

	if compareString == "flood" {
		p.AppEff = f
	} else {
		p.AppEff = s
	}
}

// FilterEff returns the efficiency of flood and sprinkler for a given year.
func FilterEff(efficiencies []database.Efficiency, year int) (f float64, s float64) {
	for _, v := range efficiencies {
		if v.Yr == year {
			f = v.AeFlood
			s = v.AeSprinkler
		}
	}

	return f, s
}
