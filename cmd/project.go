package cmd

import (
	"bitbucket.org/longitude-103/wwumm-esc/Utils"
	"bitbucket.org/longitude-103/wwumm-esc/database"
	"bitbucket.org/longitude-103/wwumm-esc/logging"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var projectOutput = &cobra.Command{
	Use:   "project",
	Short: "Projection used for Robust Review Analysis",
	Long:  "Projection command used to create a file set for robust review analysis.",
	Run: func(cmd *cobra.Command, args []string) {
		yearsStr, _ := cmd.Flags().GetString("Years")
		var years repeatYears
		err := years.parseAndSetYears(yearsStr)
		if err != nil {
			pterm.Error.Println(err)
			return
		}

		projYears, _ := cmd.Flags().GetInt("ProjectYears")
		if projYears == 0 {
			pterm.Error.Println("Projection Years must be greater than 0")
		}

		fileName, _ := cmd.Flags().GetString("FileNames")
		description, _ := cmd.Flags().GetString("Desc")

		if err := projectRR(myEnv, years, projYears, fileName, description); err != nil {
			pterm.Error.Println("Error in Application: ", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(projectOutput)
	projectOutput.Flags().String("Years", "2020", "A list of years to use for the projection, e.g. 2018, 2020, 2018 in that order, surround in double quotes")
	projectOutput.Flags().String("FileNames", "RR", "The name of the files to use for the projection output")
	projectOutput.Flags().Int("ProjectYears", 50, "The number of years to project forward")
	projectOutput.Flags().String("Desc", "Robust Review Analysis", "A description of the project run")

	_ = projectOutput.MarkFlagRequired("Years")
	_ = projectOutput.MarkFlagRequired("Desc")
}

func projectRR(myEnv map[string]string, years repeatYears, projYears int, fileName, description string) error {
	// load local database
	dbString, db, err := DbQuestion()
	if err != nil {
		return err
	}

	grid, err := database.GetGrid(db)
	if err != nil {
		return err
	}

	path, fileName, err := Utils.MakeOutputDir("qc")
	if err != nil {
		return err
	}

	v := database.Setup{}
	logFileName := fmt.Sprintf("qcFiles%s.log", fileName)
	filePath := filepath.Join(path, logFileName)

	v.PgDb, err = database.PgConnx(myEnv)
	if err != nil {
		fmt.Printf("Cannot set PgConnx")
	}

	l := logging.NewLogger(filePath)
	v.Logger = l
	v.SlDb = db

	v.Logger.Info("Using Results Database: " + dbString)

	// sqlite database
	// query each year requested, then we don't have to do it more than once
	year := years.StartYear()
	resultsData := make(map[int]database.SqliteResults)
	welResultsData := make(map[int]database.SqliteWelResults)
	spin, _ := pterm.DefaultSpinner.Start("Getting Projection Data")
	for i := 0; i < years.Len(); i++ {
		spin.UpdateText(fmt.Sprintf("Querying Year %d", year))
		// query data for this year
		res, err := database.GetResults(db, year)
		if err != nil {
			log.Printf("Error querying data: %s", err)
			return err
		}

		wRes, err := database.GetWelResults(db, year)
		if err != nil {
			log.Printf("Error querying data: %s", err)
			return err
		}

		resultsData[year] = res
		welResultsData[year] = wRes
		year = years.NextYear()
	}
	spin.Success("Projection Data Query Complete")

	// get retirement data
	spin.UpdateText("Getting Retirement Data")
	retirementData, err := database.GetRetirements(&v, grid, years.AllYears())
	if err != nil {
		log.Printf("Error getting retirement data: %s", err)
		return err
	}
	spin.Success("Retirement Data Query Complete")
	year = years.StartYear()

	spin.UpdateText("Starting Projection")
	for yr := 2021; yr < 2021+projYears; yr++ {
		spin.UpdateText("Projecting Year " + strconv.Itoa(yr))
		newRecords := make(database.SqliteResults, len(resultsData[year]))
		newWelRecords := make(database.SqliteWelResults, len(welResultsData[year]))
		copy(newRecords, resultsData[year])
		copy(newWelRecords, welResultsData[year])

		for i := 0; i < len(newRecords); i++ {
			newRecords[i].ChangeYear(yr)
		}

		for i := 0; i < len(newWelRecords); i++ {
			newWelRecords[i].ChangeYear(yr)
		}

		retireData := retirementData[year]
		retireRecords, err := retireData.SendRetirements(yr, false)
		if err != nil {
			log.Printf("Error getting retirement data: %s", err)
			return err
		}

		newWelRecords = append(newWelRecords, retireRecords...)

		if err := database.AddRCHRecords(db, newRecords); err != nil {
			return err
		}

		if err := database.AddWelRecords(db, newWelRecords); err != nil {
			return err
		}

		// which year to get
		year = years.NextYear()
	}
	spin.Success("Projection Complete")

	spin.UpdateText("Updating Output Notes")
	resultsNote, err := database.ResultsNoteDB(db)
	if err != nil {
		return err
	}

	if err := resultsNote.Add(database.Note{Nt: fmt.Sprintf("Robust Review Years: %s", years.String())}); err != nil {
		return err
	}
	if err := resultsNote.Add(database.Note{Nt: fmt.Sprintf("Robust Review Start Year: %d", 2021)}); err != nil {
		return err
	}
	if err := resultsNote.Add(database.Note{Nt: fmt.Sprintf("Robust Review End Year: %d", 2021+projYears)}); err != nil {
		return err
	}
	if err := resultsNote.Add(database.Note{Nt: fmt.Sprintf("Robust Review Description: %s", description)}); err != nil {
		return err
	}

	_ = resultsNote.Close()
	spin.Success("Output Notes Updated")

	return nil
}

type repeatYears struct {
	yearsList []int
	index     int
}

func (r *repeatYears) Len() int {
	return len(r.yearsList)
}

func (r *repeatYears) NextYear() int {
	if r.index < len(r.yearsList)-1 {
		r.index++
		yr := r.yearsList[r.index]
		return yr
	} else {
		r.index = 0
		return r.yearsList[0]
	}
}

func (r *repeatYears) StartYear() int {
	r.index = 0
	return r.yearsList[0]
}

func (r *repeatYears) parseAndSetYears(yearsStr string) error {
	var years []int
	split := strings.Split(yearsStr, ",")
	for _, yr := range split {
		yearInt, err := strconv.Atoi(strings.TrimSpace(yr))
		if err != nil {
			return errors.New("error parsing years: " + err.Error())
		}

		years = append(years, yearInt)
	}

	if len(years) == 0 {
		return errors.New("must include at least one year")
	}

	r.yearsList = years

	return nil
}

func (r *repeatYears) String() string {
	if len(r.yearsList) == 0 {
		return ""
	}

	yearsString := strconv.Itoa(r.yearsList[0])
	for i := 1; i < len(r.yearsList); i++ {
		yearsString += ", " + strconv.Itoa(r.yearsList[i])
	}

	return yearsString
}

func (r *repeatYears) AllYears() []int {
	return r.yearsList
}
