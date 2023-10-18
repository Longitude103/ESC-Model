package conveyLoss

import (
	"fmt"
	"os"
	"testing"

	"bitbucket.org/longitude-103/wwumm-esc/database"
	"github.com/joho/godotenv"
)

func dbConnection() *database.Setup {
	var myEnv map[string]string
	myEnv, err := godotenv.Read("../../.env")
	if err != nil {
		fmt.Println("Cannot load Env Variables:", err)
		os.Exit(1)
	}

	var v *database.Setup
	v, err = database.NewSetup(myEnv, database.WithNoSQLite(), database.WithDebug())
	if err != nil {
		fmt.Println("Error in NewSetup: ", err)
	}

	if err = v.SetYears(1997, 1997); err != nil {
		fmt.Println("error setting years")
	}

	return v
}

func Test_getCanals(t *testing.T) {
	v := dbConnection()
	v.SYear = 2014
	v.EYear = 2016

	c, err := getCanals(v)
	if err != nil {
		t.Errorf("Error getting canals: %s", err)
	}

	foundStartYr := false
	foundEndYr := false
	for _, canal := range c {
		v.Logger.Debugf("Canal: %+v\n", canal)
		if canal.Yr == v.SYear {
			foundStartYr = true
		}
		if canal.Yr == v.EYear {
			foundEndYr = true
		}
	}

	if !foundStartYr {
		t.Errorf("Didn't find any canals from start year of %d", v.SYear)
	}

	if !foundEndYr {
		t.Errorf("Didn't find any canals from end year of %d", v.EYear)
	}

	if len(c) == 0 {
		t.Error("Didn't return any canals")
	}

}

func Test_getCanalsSS(t *testing.T) {
	v := dbConnection()
	v.SYear = 1895
	v.EYear = 1905
	v.SteadyState = true

	v.Logger.Infof("Start Year %d", v.SYear)
	v.Logger.Infof("End Year %d", v.EYear)

	c, err := getCanals(v)
	if err != nil {
		t.Errorf("Error getting canals: %s", err)
	}

	foundStartYr := false
	foundEndYr := false
	for _, canal := range c {
		v.Logger.Debugf("Canal: %+v\n", canal)
		if canal.Yr == v.SYear {
			foundStartYr = true
		}
		if canal.Yr == v.EYear {
			foundEndYr = true
		}
	}

	if !foundStartYr {
		t.Errorf("Didn't find any canals from start year of %d", v.SYear)
	}

	if !foundEndYr {
		t.Errorf("Didn't find any canals from end year of %d", v.EYear)
	}

	if len(c) == 0 {
		t.Error("Didn't return any canals")
	}
}

func Test_getCanalCells(t *testing.T) {
	v := dbConnection()

	cc, err := getCanalCells(v)
	if err != nil {
		t.Errorf("Error getting canal cells: %s", err)
	}

	for _, cell := range cc {
		if cell.CLinkId == 25 {
			v.Logger.Debugf("Canal 25: %+v", cell)
		}

	}

	if len(cc) == 0 {
		t.Error("Didn't return any canal cells")
	}
}
