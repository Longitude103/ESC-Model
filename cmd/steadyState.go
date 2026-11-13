package cmd

import (
	"fmt"
	"github.com/Longitude103/ESC-Model/parcels"
	"github.com/spf13/cobra"
	"os"
	"time"

	"github.com/Longitude103/ESC-Model/database"
	"github.com/Longitude103/ESC-Model/fileio"
	"github.com/Longitude103/ESC-Model/rchFiles"
	"github.com/pterm/pterm"
)

var runSteadyStateCmd = &cobra.Command{
	Use:   "runSteadyState",
	Short: "Run the EscModel in steady state mode",
	Long:  `This command is the steady state model execution for the EscModel. This function will run the model using the provided configuration for steady state. It will create a new output database file on your local filesystem in the same directory as the executable.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("runModel called")

		sYr, _ := cmd.Flags().GetInt("StartYr")
		eYr, _ := cmd.Flags().GetInt("EndYr")
		asYr, _ := cmd.Flags().GetInt("AvgStart")
		aeYr, _ := cmd.Flags().GetInt("AvgEnd")
		if aeYr == 0 {
			aeYr = configEndYear
		}
		csDir, _ := cmd.Flags().GetString("CSDir")
		desc, _ := cmd.Flags().GetString("Desc")
		oldGrid, _ := cmd.Flags().GetBool("oldGrid")
		mf6Grid40, _ := cmd.Flags().GetBool("mf6Grid40")

		if oldGrid && mf6Grid40 {
			pterm.Error.Println("oldGrid and mf6Grid40 are mutually exclusive")
			return
		}

		if err := runSteadyState(desc, csDir, sYr, eYr, asYr, aeYr, oldGrid, mf6Grid40, myEnv); err != nil {
			pterm.Error.Printf("Error in Application: %s\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runSteadyStateCmd)
	runSteadyStateCmd.Flags().IntP("StartYr", "s", 1893, "Sets the start year of Command")
	runSteadyStateCmd.Flags().IntP("EndYr", "e", 1952, "Sets the end year of Command")
	runSteadyStateCmd.Flags().IntP("AvgStart", "a", 1953, "Sets the start year of Averaging")
	runSteadyStateCmd.Flags().IntP("AvgEnd", "q", 0, "Sets the end year of Averaging (default: endYear from the config file)")
	runSteadyStateCmd.Flags().StringP("CSDir", "c", "", "REQUIRED CropSim Directory path")
	runSteadyStateCmd.Flags().StringP("Desc", "d", "", "REQUIRED Model Description")
	runSteadyStateCmd.Flags().BoolP("oldGrid", "o", false, "If flag set, the model will use the 40 acre grid, not USG as default")
	runSteadyStateCmd.Flags().BoolP("mf6Grid40", "m", false, "If flag set, the model will use the 40 acre grid but in MF6 Node Numbers")
	_ = runSteadyStateCmd.MarkFlagRequired("CSDir")
	_ = runSteadyStateCmd.MarkFlagRequired("Desc")
}

// RunSteadyState is a function that runs the model in Steady State Mode. This produces the following recharge file, but no
// well file is produced.
func runSteadyState(mDesc, CSDir string, StartYr, EndYr, AvgStart, AvgEnd int, oldGrid, mf640 bool, myEnv map[string]string) error {
	// first stress period is all Natural Veg for whole grid
	// Second stress period is all natural veg and a repeat of the 1st period
	// Third -> end is a monthly stress periods using surface water data starting January 1895 to December 1952
	timeStart := time.Now()

	pterm.Info.Println("Setting up results database")
	var opts []database.Option
	opts = append(opts, database.WithSteadyState(StartYr, EndYr))
	opts = append(opts, database.WithAverageYears(AvgStart, AvgEnd))

	if oldGrid {
		opts = append(opts, database.WithOldGrid())
	}

	if mf640 {
		opts = append(opts, database.WithMF640Grid())
	}

	v, err := database.NewSetup(myEnv, opts...)
	if err != nil {
		return err
	}

	v.Logger.Infof("Model Run Started at: %s", timeStart.Format(time.UnixDate))
	pterm.Info.Printf("Model Description: %s\n", mDesc)
	v.Logger.Infof("Model Description: %s", mDesc)

	noteDb, err := database.ResultsNoteDB(v.SlDb)
	if err != nil {
		return err
	}

	avgStartYearNote := fmt.Sprintf("Average Start Year: %d", AvgStart)
	avgEndYearNote := fmt.Sprintf("Average End Year: %d", AvgEnd)
	sYearNote := fmt.Sprintf("Start Year: %d", v.SYear)
	eYearNote := fmt.Sprintf("End Year: %d", v.EYear)

	if err := noteDb.Add(database.Note{Nt: sYearNote}); err != nil {
		return err
	}
	if err := noteDb.Add(database.Note{Nt: eYearNote}); err != nil {
		return err
	}
	if err := noteDb.Add(database.Note{Nt: "Desc: " + mDesc}); err != nil {
		return err
	}
	if err := noteDb.Add(database.Note{Nt: avgStartYearNote}); err != nil {
		return err
	}
	if err := noteDb.Add(database.Note{Nt: avgEndYearNote}); err != nil {
		return err
	}
	if err := noteDb.Add(database.Note{Nt: "SteadyState Model Run"}); err != nil {
		return err
	}

	if err := database.AddCellsToOutput(v); err != nil {
		return err
	}
	if v.OldGrid {
		if err := noteDb.Add(database.Note{Nt: "grid=1"}); err != nil {
			return err
		}
	} else {
		if err := noteDb.Add(database.Note{Nt: "grid=2"}); err != nil {
			return err
		}
	}

	// get cropsim files loaded
	spinnerSuccess, _ := pterm.DefaultSpinner.Start("Reading CropSim Results files")
	csResults, err := fileio.LoadTextFiles(CSDir, v.Logger)
	if err != nil {
		spinnerSuccess.Fail("Error in Loading Text Files")
		return err
	}

	v.Logger.Info("Averaging CS Results")
	avgCSResults, err := fileio.AverageStationResults(csResults, AvgStart, AvgEnd)
	if err != nil {
		spinnerSuccess.Fail("Error in Averaging CS Results")
		return err
	}

	spinnerSuccess.Success()

	// get Weather Stations
	spinnerSuccess, _ = pterm.DefaultSpinner.Start("Getting Weather Stations")
	v.Logger.Info("Getting Weather Stations")
	wStations, err := database.GetWeatherStations(v.PgDb)
	if err != nil {
		spinnerSuccess.Fail("Error Getting Weather Stations")
		return err
	}
	spinnerSuccess.Success()

	spinnerSuccess, _ = pterm.DefaultSpinner.Start("Getting Coefficients of Crops")
	v.Logger.Info("Getting Coefficients of Crops")
	cCoefficients, err := database.GetCoeffCrops(v)
	if err != nil {
		return err
	}
	spinnerSuccess.Success()

	// create average natural vegetation values for each month
	v.Logger.Info("Preforming Natural Vegetation Calculations")
	if err := rchFiles.NaturalVegSS(v, wStations, avgCSResults, cCoefficients); err != nil {
		v.Logger.Errorf("Error in Natural Vegetation: %s", err)
		return err
	}

	// parcel pumping
	v.Logger.Info("Preforming Parcel Pumping")
	irrParcels, err := parcels.ParcelPumpSS(v, avgCSResults, wStations, cCoefficients)
	if err != nil {
		v.Logger.Errorf("Error in Parcel Pumping: %s", err)
		return err
	}
	pterm.Success.Println("Successfully Completed Parcel Pumping")

	if err := v.PNirDB.Flush(); err != nil {
		v.Logger.Errorf("Error in flush: %s", err)
	}

	// Irr Cells
	v.Logger.Info("Preforming Irrigation RCH Operations")
	if err := rchFiles.IrrigationRCH(v, irrParcels, cCoefficients); err != nil {
		v.Logger.Errorf("Error in Creating Irrigation RCH %s", err)
		return err
	}
	pterm.Success.Println("Successfully Completed Irrigated Results")
	if err := v.RchDb.Close(); err != nil {
		return err
	}

	v.Logger.Info("Preforming Dryland Parcel Operations")
	dryParcels, err := parcels.DryLandParcels(v, avgCSResults, wStations, cCoefficients)
	if err != nil {
		v.Logger.Errorf("Error in Dry Land Parcels: %s", err)
	}
	pterm.Success.Println("Successfully Completed Dryland Parcel Ops")

	// Dryland 101
	if err := rchFiles.Dryland(v, dryParcels, cCoefficients); err != nil {
		v.Logger.Errorf("Error in Dryland: %s", err)
		return err
	}
	pterm.Success.Println("Successfully Completed Dryland Results")

	if err := v.RchDb.Flush(); err != nil {
		return err
	}

	_ = noteDb.Close()
	_ = v.SlDb.Close() // close the db before ending the program
	v.Logger.Infof("Model Runtime: %s", time.Since(timeStart))
	v.Logger.Info("Steady State Model Completed Normally")
	pterm.Info.Println("Steady State Model Completed Normally, check logs for details of run")
	v.Logger.Close()
	return nil
}
