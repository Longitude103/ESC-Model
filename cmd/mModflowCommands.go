package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"

	"github.com/spf13/cobra"

	"github.com/pterm/pterm"

	Flogo "bitbucket.org/longitude-103/flogo"
	"bitbucket.org/longitude-103/wwumm-esc/Utils"
	"bitbucket.org/longitude-103/wwumm-esc/database"
	"github.com/jmoiron/sqlx"
)

var makeModflowFilesCmd = &cobra.Command{
	Use:   "mfFiles",
	Short: "Create ModFlow files",
	Long:  `Create ModFlow files for use in the modflow application. It will create a new wel and rch file that can be directly imported. The function will ask a series of questions to determine which files to create.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := makeModflowFiles(); err != nil {
			pterm.Error.Println("Error in Application: ", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(makeModflowFilesCmd)
}

func makeModflowFiles() error {
	_, db, err := DbQuestion()
	if err != nil {
		return err
	}

	path, _, err := Utils.MakeOutputDir("mfFiles")
	if err != nil {
		return err
	}

	steadyState, err := database.GetSteadyState(db)
	if err != nil {
		return err
	}

	mDesc, err := database.GetDescription(db)
	if err != nil {
		return err
	}

	StartYear, EndYear, err := database.GetStartEndYrs(db)
	if err != nil {
		return err
	}

	startDate := time.Date(StartYear, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(EndYear, 12, 1, 0, 0, 0, 0, time.UTC)

	a, err := questions(db, steadyState)
	if err != nil {
		return err
	}

	var aggWel, aggRch []database.MfResults

	if steadyState {
		pterm.Info.Println("Steady State Run Detected, will only make RCH File.")
		pterm.Info.Println("Processing RCH from Local DB")
		aggRch, err = database.GetAggResults(db, false, []string{})
		if err != nil {
			return err
		}

		if err := processSSAggRCH(aggRch, "AggregateRCH", path, mDesc, startDate, endDate); err != nil {
			return err
		}

		pterm.Success.Println("Created RCH File, look in OutputFiles directory")
		return nil
	}

	if !steadyState {
		aggWel, err = database.GetAggResults(db, true, a.WellFK)
		if err != nil {
			return err
		}

		var singleWELResults = make(map[string][]database.MfResults)
		var singleRCHResults = make(map[string][]database.MfResults)

		for _, w := range a.WellFK {
			singleWELResults[w], err = database.SingleResult(db, true, w)
			if err != nil {
				return err
			}
		}

		for _, r := range a.RchFK {
			singleRCHResults[r], err = database.SingleResult(db, false, r)
			if err != nil {
				return err
			}
		}

		if err := MakeFiles(aggWel, true, false, a.RowCol, "AggregateWEL", path, mDesc, startDate, endDate); err != nil {
			return err
		}

		for k := range singleWELResults {
			if len(singleWELResults[k]) > 0 {
				noSpaces := strings.Replace(k, " ", "_", -1)
				fn := fmt.Sprintf("%sWEL", noSpaces)
				if err := MakeFiles(singleWELResults[k], true, false, a.RowCol, fn, path, mDesc, startDate, endDate); err != nil {
					return err
				}
			}
		}

		for k := range singleRCHResults {
			if len(singleRCHResults[k]) > 0 {
				noSpaces := strings.Replace(k, " ", "_", -1)
				fn := fmt.Sprintf("%sRCH", noSpaces)
				if err := MakeFiles(singleRCHResults[k], false, true, a.RowCol, fn, path, mDesc, startDate, endDate); err != nil {
					return err
				}
			}
		}

		aggRch, err = database.GetAggResults(db, false, a.RchFK)
		if err != nil {
			return err
		}

		if err := MakeFiles(aggRch, false, true, a.RowCol, "AggregateRCH", path, mDesc, startDate, endDate); err != nil {
			return err
		}
	}

	return nil
}

func suggestFiles(toComplete string) []string {
	files, _ := filepath.Glob("./OutputFiles/*/*.sqlite")
	return files
}

func DbQuestion() (string, *sqlx.DB, error) {
	var q = []*survey.Question{
		{
			Name: "file",
			Prompt: &survey.Input{
				Message: "Which results DB should be used?",
				Suggest: suggestFiles,
				Help:    "The SQLite file you want to use to build ModFlow files",
			},
			Validate: survey.Required,
		},
	}

	answers := struct {
		File string
	}{}

	// ask the question
	if err := survey.Ask(q, &answers); err != nil {
		return "", nil, err
	}

	// get the file_keys from sqlite and ask if any of these should be individual files
	sqliteDB, err := database.ConnectSqlite(answers.File)
	if err != nil {
		return "", nil, err
	}

	return answers.File, sqliteDB, nil
}

type Answers struct {
	WellFK []string
	RchFK  []string
	RowCol bool
}

func questions(sqliteDB *sqlx.DB, steadyState bool) (Answers, error) {
	var multiQs []*survey.Question
	if steadyState {
		multiQs = []*survey.Question{
			{
				Name: "RowCol",
				Prompt: &survey.Confirm{
					Message: "Do you want to Output Row-Column (node is default)?",
				},
			},
		}
	} else {
		wellFk, err := database.GetFileKeys(sqliteDB, true)
		if err != nil {
			return Answers{}, err
		}

		rchFK, err := database.GetFileKeys(sqliteDB, false)
		if err != nil {
			return Answers{}, err
		}

		// the questions to ask
		multiQs = []*survey.Question{
			{
				Name: "WellFK",
				Prompt: &survey.MultiSelect{
					Message: "Choose WEL files to exclude :",
					Options: wellFk,
				},
			},
			{
				Name: "RchFK",
				Prompt: &survey.MultiSelect{
					Message: "Choose which RCH files to exclude :",
					Options: rchFK,
				},
			},
			{
				Name: "RowCol",
				Prompt: &survey.Confirm{
					Message: "Do you want to Output Row-Column (node is default)?",
				},
			},
		}
	}

	var a Answers
	// ask the questions
	if err := survey.Ask(multiQs, &a); err != nil {
		return Answers{}, err
	}

	return a, nil
}

func MakeFiles(r []database.MfResults, wel bool, rch bool, Rc bool, fileName string, outputPath string, mDesc string, modelStartDate, modelEndDate time.Time) error {
	rInterface := make([]interface {
		Date() time.Time
		Node() int
		Value() float64
		UseValue() bool
		RowCol() (int, int)
		ConvertToFtPDay() float64
		ConvertToFt3PDay() float64
	}, len(r))
	for i, v := range r {
		if wel {
			// acre-feet / month -> (ft^3 / day) * -1
			v.Rslt = (v.Rslt * 43560) / float64(Utils.TimeExt{T: v.ResultDate}.DaysInMonth()) * -1
			v.SetConvertedValue()
		}

		if rch {
			// acre-feet / month -> ft / day
			v.Rslt = (v.Rslt / v.CellSize.Float64) / float64(Utils.TimeExt{T: v.ResultDate}.DaysInMonth())
			v.SetConvertedValue()
		}
		rInterface[i] = v
	}

	if err := Flogo.Input(wel, rch, Rc, fileName, rInterface, outputPath, mDesc, modelStartDate, modelEndDate); err != nil {
		return err
	}

	return nil
}

func processSSAggRCH(results SliceMfResults, fileName, outputPath, mDesc string, modelStartDate, modelEndDate time.Time) error {
	Rc := true
	if results[0].IsNodeResult() {
		Rc = false
	}

	pterm.Info.Println("Making First two time period annual results for SS Model")
	annualResults := filterResultsForYear(results)

	pterm.Info.Println("Excluding the annual years")
	excludedResults := results.ExcludeResults([]int{1893, 1894})
	allResults := append(annualResults, excludedResults...)

	rInterface := make([]interface {
		Date() time.Time
		Node() int
		Value() float64
		UseValue() bool
		RowCol() (int, int)
		ConvertToFtPDay() float64
		ConvertToFt3PDay() float64
	}, len(allResults))

	for i := 0; i < len(allResults); i++ {
		rInterface[i] = allResults[i]
	}

	pterm.Info.Println("Creating RCH File")
	if err := Flogo.Input(false, true, Rc, fileName, rInterface, outputPath, mDesc, modelStartDate, modelEndDate); err != nil {
		return err
	}

	return nil
}

type SliceMfResults []database.MfResults

// filterMyResults is a method to filter SliceMfResults slice of database.MfResults into just the results for the year passed
// to the method.
func (mr SliceMfResults) filterMyResults(yr int) SliceMfResults {
	var mnthResults []database.MfResults
	for _, r := range mr {
		if r.Year() == yr {
			mnthResults = append(mnthResults, r)
		}
	}

	return mnthResults
}

// GroupToAnnual is a method of SliceMfResults that will group all items within the slice by node. Make sure there are only
// one year of data when calling this method.
func (mr SliceMfResults) GroupToAnnual() map[int]SliceMfResults {
	resultMap := make(map[int]SliceMfResults)
	for _, r := range mr {
		resultMap[r.Node()] = append(resultMap[r.Node()], r)
	}

	return resultMap
}

// ExcludeResults is a method to return a subset of SliceMfResults that does not have a slice of years present.
func (mr SliceMfResults) ExcludeResults(yrs []int) SliceMfResults {
	var result SliceMfResults

	for _, r := range mr {
		found := false
		for i := 0; i < len(yrs); i++ {
			if r.Year() == yrs[i] {
				found = true
			}
		}

		if found { // found a year, don't amend
			continue
		} else {
			result = append(result, r)
		}
	}

	return result
}

// filterResultsForYear is a function for steady state to convert the first two years of data into two annual datasets and
// return them in a new slice where they are in the struct with dates of 11/1894 and 12/1894, but are annual datasets for the
// first two time steps of the steady state model.
func filterResultsForYear(allResults SliceMfResults) SliceMfResults {
	var results SliceMfResults
	for _, yr := range []int{1893, 1894} {
		mnthResult := allResults.filterMyResults(yr)
		groupedResults := mnthResult.GroupToAnnual()

		for k, listResults := range groupedResults {
			var totalAF float64
			for _, lr := range listResults {
				totalAF += lr.Rslt
			}

			// acre-feet / year -> ft / day, cell size cannot change during model
			Rslt := (totalAF / listResults[0].CellSize.Float64) / float64(Utils.TimeExt{T: listResults[0].ResultDate}.DaysInYear())

			mnth := 11
			if yr == 1894 {
				mnth = 12
			}

			mfResult := database.MfResults{ResultDate: time.Date(1894, time.Month(mnth), 1, 0, 0, 0, 0, time.UTC),
				CellNode: k, Rslt: Rslt, CellSize: listResults[0].CellSize, Rw: listResults[0].Rw, Clm: listResults[0].Clm, ConvertedValue: true}

			results = append(results, mfResult)
		}

	}

	return results
}
