package wells

import "github.com/Longitude103/ESC-Model/database"

func AddBackRetirements(v *database.Setup) error {
	startYear := 1998
	if v.SYear > startYear {
		startYear = v.SYear
	}

	for y := startYear; y < v.EYear+1; y++ {
		retirements, err := database.GetOneYearRetirements(v, y)
		if err != nil {
			v.Logger.Errorf("Error getting retirements for year %d: %s", y, err)
			return err
		}

		for i := 0; i < len(retirements); i++ {
			if err := database.AddWelRecords(v.SlDb, retirements[i].ToSqliteWelResults(y, false)); err != nil {
				v.Logger.Errorf("Error adding retirements for year %d: %s", y, err)
				continue
			}
		}
	}

	return nil
}
