package parcels

import (
	"fmt"

	"bitbucket.org/longitude-103/wwumm-esc/database"
)

type Usage struct {
	CertNum string  `db:"cert_num"`
	UseAF   float64 `db:"usage_af"`
	Yr      int     `db:"yr"`
	Nrd     string  `db:"nrd"`
}

// GetUsage function returns a map with a year of year and a value of slice of usage struct for all cert usage in both nrds. It includes all years.
func GetUsage(v *database.Setup) map[int][]Usage {
	query := `SELECT cert_num::varchar, usage_af, yr, 'np' nrd from np.np_usage UNION SELECT cert_num, usage_af, yr, 'sp' nrd from sp.sp_usage;`

	var use []Usage
	err := v.PgDb.Select(&use, query)
	if err != nil {
		fmt.Println("Error in Usage:", err)
	}

	useMap := make(map[int][]Usage)

	for Yr := v.SYear; Yr < v.EYear+1; Yr++ {
		useMap[Yr] = filterUsage(use, Yr)
	}

	return useMap
}

// filterUsage is a function that filters the total usage to a year and returns a slice of Usage
func filterUsage(u []Usage, yr int) (filteredUsage []Usage) {
	for _, v := range u {
		if v.Yr == yr {
			filteredUsage = append(filteredUsage, v)
		}
	}

	return filteredUsage
}

// DistributeUsage method receives the total NIR, monthly NIR values and usage for the cert and distributes that pumping
// to the parcel. It also sets the parcel metered property to true.
func (p *Parcel) DistributeUsage(totalNIR float64, totalMonthlyNIR [12]float64, u float64) {
	for m := 0; m < 12; m++ {
		if totalMonthlyNIR[m] > 0 { // protect from division by zero
			monthPercent := totalMonthlyNIR[m] / totalNIR
			p.Pump[m] = p.Nir[m] / totalMonthlyNIR[m] * (monthPercent * u)
		}
	}
}
