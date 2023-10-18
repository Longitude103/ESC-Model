package cmd

import (
	"fmt"
	"log"

	"bitbucket.org/longitude-103/wwumm-esc/database"
	"github.com/spf13/cobra"
)

var parcelDeleteCmd = &cobra.Command{
	Use:   "parcel-delete",
	Short: "Delete parcel",
	Long:  `Delete a parcel along with the junction table values within that parcel.`,
	Run: func(cmd *cobra.Command, args []string) {
		yr, _ := cmd.Flags().GetInt("year")
		parcelId, _ := cmd.Flags().GetInt("id")
		if parcelId == 0 {
			fmt.Println("please provide a parcel id")
			return
		}

		nrd, _ := cmd.Flags().GetString("nrd")
		if nrd != "np" && nrd != "sp" {
			fmt.Println("please provide a valid nrd, must be 'np' or 'sp'")
			return
		}

		// get database
		var opts []database.Option
		opts = append(opts, database.WithNoSQLite())
		v, err := database.NewSetup(myEnv, opts...)
		if err != nil {
			log.Fatal("Failed to create database: ", err)
		}

		// look for parcel and tell them if we don't find it
		query := fmt.Sprintf("select EXISTS(select parcel_id from %s.t%d_irr where parcel_id = %d);", nrd, yr, parcelId)

		var exists bool
		if err := v.PgDb.Get(&exists, query); err != nil {
			log.Fatal(err)
		}

		if !exists {
			fmt.Println("Parcel not found")
			return
		}

		// delete data in jct table
		query = fmt.Sprintf("delete from %s.t%d_jct where parcel_id = %d;", nrd, yr, parcelId)
		if _, err := v.PgDb.Exec(query); err != nil {
			log.Fatal(err)
		}

		// delete data in parcel table
		query = fmt.Sprintf("delete from %s.t%d_irr where parcel_id = %d;", nrd, yr, parcelId)
		if _, err := v.PgDb.Exec(query); err != nil {
			log.Fatal(err)
		}

		// tell the user we're done
		fmt.Println("Deleted parcel with id ", parcelId)
	},
}

func init() {
	rootCmd.AddCommand(parcelDeleteCmd)
	parcelDeleteCmd.Flags().Int("year", 2020, "Year of Parcel to Delete")
	parcelDeleteCmd.Flags().Int("id", 0, "ID of Parcel to Delete")
	parcelDeleteCmd.Flags().String("nrd", "", "NRD of Parcel to Delete")

	parcelDeleteCmd.MarkFlagRequired("year")
	parcelDeleteCmd.MarkFlagRequired("id")
	parcelDeleteCmd.MarkFlagRequired("nrd")
}
