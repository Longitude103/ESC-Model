package cmd

import (
	"bitbucket.org/longitude-103/wwumm-esc/database"
	"errors"
	"fmt"
	"log"
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

		if err := projectRR(myEnv, years, projYears, fileName); err != nil {
			pterm.Error.Println("Error in Application: ", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(projectOutput)
	projectOutput.Flags().String("Years", "2020", "A list of years to use for the projection, e.g. 2018, 2020, 2018 in that order, surround in double quotes")
	projectOutput.Flags().String("FileNames", "RR", "The name of the files to use for the projection output")
	projectOutput.Flags().Int("ProjectYears", 50, "The number of years to project forward")

	_ = projectOutput.MarkFlagRequired("Years")
}

func projectRR(myEnv map[string]string, years repeatYears, projYears int, fileName string) error {
	// load local database
	_, db, err := DbQuestion()
	if err != nil {
		return err
	}

	// sqlite database
	// query each year requested, then we don't have to do it more than once once
	year := years.StartYear()
	resultsData := make(map[int]database.SqliteResults)
	welResultsData := make(map[int]database.SqliteWelResults)
	for i := 0; i < years.Len(); i++ {
		// query data for this year
		fmt.Printf("Querying data for %d\n", year)
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

		fmt.Printf("Results have %d records\n", len(res))

		resultsData[year] = res
		welResultsData[year] = wRes
		year = years.NextYear()
		fmt.Printf("Next year is %d\n", year)
	}

	year = years.StartYear()
	for yr := 2021; yr < 2021+projYears; yr++ {
		fmt.Printf("Projecting data for %d\n", yr)
		fmt.Printf("Using year: %d\n", year)
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

		if err := database.AddRCHRecords(db, newRecords); err != nil {
			return err
		}

		if err := database.AddWelRecords(db, newWelRecords); err != nil {
			return err
		}

		// which year to get
		year = years.NextYear()
	}

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
