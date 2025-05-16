package cmd

import (
	"bitbucket.org/longitude-103/wwumm-esc/database"
	"database/sql"
	"fmt"
	"github.com/spf13/cobra"
	"log"
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

var jctDeleteCmd = &cobra.Command{
	Use:   "jct-delete",
	Short: "Delete JCT record",
	Long:  `Delete a junction table record by itself.`,
	Run: func(cmd *cobra.Command, args []string) {
		yr, _ := cmd.Flags().GetInt("year")
		parcelId, _ := cmd.Flags().GetInt("parcel_id")
		if parcelId == 0 {
			fmt.Println("please provide a parcel id")
			return
		}

		wellId, _ := cmd.Flags().GetInt("well_id")
		if wellId == 0 {
			fmt.Println("please provide a well id")
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

		// look for jct record and tell them if we don't find it
		query := fmt.Sprintf("select EXISTS(select parcel_id, wellid from %s.t%d_jct where parcel_id = %d and wellid = %d);", nrd, yr, parcelId, wellId)

		var exists bool
		if err := v.PgDb.Get(&exists, query); err != nil {
			log.Fatal(err)
		}

		if !exists {
			fmt.Println("Junction Table record not found")
			return
		}

		// delete data in jct table
		query = fmt.Sprintf("delete from %s.t%d_jct where parcel_id = %d and wellid = %d;", nrd, yr, parcelId, wellId)
		if _, err := v.PgDb.Exec(query); err != nil {
			log.Fatal(err)
		}

		// tell the user we're done
		fmt.Printf("Deleted junction table record with parcel id of %d and well id of %d", parcelId, wellId)
	},
}

var parcelCopyCmd = &cobra.Command{
	Use:   "parcel-copy",
	Short: "copy parcel from one year to another",
	Long:  `Will copy a parcel from one year to another, preserving the junction table values.`,
	Run: func(cmd *cobra.Command, args []string) {
		yrTo, _ := cmd.Flags().GetInt("year_to")
		yrFrom, _ := cmd.Flags().GetInt("year_from")
		parcelId, _ := cmd.Flags().GetInt("parcel_id")
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
		query := fmt.Sprintf("select EXISTS(select parcel_id from %s.t%d_irr where parcel_id = %d);", nrd, yrFrom, parcelId)

		var exists bool
		if err := v.PgDb.Get(&exists, query); err != nil {
			log.Fatal(err)
		}

		if !exists {
			fmt.Println("Parcel not found")
			return
		}

		type Parcel struct {
			ParcelNo       int             `db:"parcel_id"`
			CertNum        sql.NullString  `db:"cert_num"`
			Crop1          sql.NullString  `db:"crop1"`
			Crop2          sql.NullString  `db:"crop2"`
			Crop3          sql.NullString  `db:"crop3"`
			Crop4          sql.NullString  `db:"crop4"`
			IrrType        sql.NullString  `db:"irrig_type"`
			FirstIrr       sql.NullInt64   `db:"first_irr"`
			SwFac          sql.NullString  `db:"sw_fac"`
			ModelId        sql.NullString  `db:"model_id"`
			Crop1Cov       sql.NullFloat64 `db:"crop1_cov"`
			Crop2Cov       sql.NullFloat64 `db:"crop2_cov"`
			Crop3Cov       sql.NullFloat64 `db:"crop3_cov"`
			Crop4Cov       sql.NullFloat64 `db:"crop4_cov"`
			Sw             sql.NullBool    `db:"sw"`
			Gw             sql.NullBool    `db:"gw"`
			Subarea        sql.NullString  `db:"subarea"`
			Oa             sql.NullInt64   `db:"oa"`
			SwID           sql.NullInt64   `db:"sw_id"`
			Objectid       sql.NullInt64   `db:"objectid"`
			County         sql.NullString  `db:"county"`
			Comments       sql.NullString  `db:"comments"`
			OldCertNum     sql.NullString  `db:"old_cert_num"`
			NewCertNum     sql.NullString  `db:"new_cert_num"`
			RecertComments sql.NullString  `db:"recert_comments"`
			NpnrdNoUsage   sql.NullString  `db:"npnrd_no_usage"`
			NpnrdFlag      sql.NullString  `db:"npnrd_flag"`
			CreatedUser    sql.NullString  `db:"created_user"`
			CreatedDate    sql.NullTime    `db:"created_date"`
			LastEditedUser sql.NullString  `db:"last_edited_user"`
			LastEditedDate sql.NullTime    `db:"last_edited_date"`
			Geom           sql.NullString  `db:"geom"`
			CalYear        sql.NullInt64   `db:"cal_year"`
			Acres          sql.NullFloat64 `db:"acres"`
			UrfId          sql.NullInt64   `db:"urf_id"`
			CrpSrc         sql.NullString  `db:"crp_src"`
			Notes          sql.NullString  `db:"notes"`
			Mod1           sql.NullString  `db:"mod1"`
		}

		// Copy parcel data
		query = fmt.Sprintf(`select objectid, irrig_type, sw_fac, county, subarea, comments, crop1, crop2, crop3, 
       crop4, old_cert_num, new_cert_num, recert_comments, npnrd_no_usage, oa, npnrd_flag, created_user, created_date, 
       last_edited_user, last_edited_date, geom, cal_year, parcel_id, acres, sw, gw, cert_num, first_irr, urf_id, crop1_cov, 
       crop2_cov, crop3_cov, crop4_cov, crp_src, model_id, notes, sw_id, mod1 from %s.t%d_irr where parcel_id = %d;`, nrd, yrFrom, parcelId)

		var parcel Parcel
		if err := v.PgDb.Get(&parcel, query); err != nil {
			log.Fatal(err)
		}

		type Jct struct {
			ParcelId int `db:"parcel_id"`
			WellId   int `db:"wellid"`
		}

		// Copy junction table data
		query = fmt.Sprintf(`select * from %s.t%d_jct where parcel_id = %d;`, nrd, yrTo, parcel.ParcelNo)

		var jcts []Jct
		if err := v.PgDb.Select(&jcts, query); err != nil {
			log.Fatal(err)
		}

		// look for the parcel id in the new year table records
		// look for parcel and tell them if we don't find it
		query = fmt.Sprintf("select EXISTS(select parcel_id from %s.t%d_irr where parcel_id = %d);", nrd, yrTo, parcelId)
		if err := v.PgDb.Get(&exists, query); err != nil {
			log.Fatal(err)
		}

		if exists {
			fmt.Println("Parcel Id already exists in the new year table records")

			// find the max parcel id in the new year table records
			query = fmt.Sprintf("select MAX(parcel_id) from %s.t%d_irr;", nrd, yrTo)
			var maxParcelId int
			if err := v.PgDb.Get(&maxParcelId, query); err != nil {
				log.Fatal(err)
			}

			parcel.ParcelNo = maxParcelId + 1
			// update the new parcel id in the junction table data
			for i := range jcts {
				jcts[i].ParcelId = parcel.ParcelNo
			}
		}

		// insert the copy data into the new year table records
		query = fmt.Sprintf(`INSERT INTO %s.t%d_irr (
			parcel_id, cert_num, crop1, crop2, crop3, crop4, irrig_type, first_irr, sw_fac, model_id,
			crop1_cov, crop2_cov, crop3_cov, crop4_cov, sw, gw, subarea, oa, sw_id,
			objectid, county, comments, old_cert_num,
			new_cert_num, recert_comments, npnrd_no_usage, npnrd_flag, created_user, created_date,
			last_edited_user, last_edited_date, geom, cal_year, acres, urf_id, crp_src, notes, mod1
		) VALUES (
			:parcel_id, :cert_num, :crop1, :crop2, :crop3, :crop4, :irrig_type, :first_irr, :sw_fac, :model_id,
			:crop1_cov, :crop2_cov, :crop3_cov, :crop4_cov, :sw, :gw, :subarea, :oa, :sw_id,
			:objectid, :county, :comments, :old_cert_num,
			:new_cert_num, :recert_comments, :npnrd_no_usage, :npnrd_flag, :created_user, :created_date,
			:last_edited_user, :last_edited_date, :geom, :cal_year, :acres, :urf_id, :crp_src, :notes, :mod1
		);`, nrd, yrTo)

		// NamedExec uses the db tags to map struct fields to query parameters
		_, err = v.PgDb.NamedExec(query, parcel)
		if err != nil {
			log.Fatal(err)
		}

		// update the junction table data with the new parcel id if it was found in the new year table records
		query = fmt.Sprintf(`INSERT INTO %s.t%d_jct (parcel_id, wellid) VALUES (:parcel_id, :wellid);`, nrd, yrTo)

		for _, jct := range jcts {
			_, err = v.PgDb.NamedExec(query, jct)
			if err != nil {
				log.Fatal(err)
			}
		}

		// tell the user we're done
		fmt.Printf("Successfully added new parcel with Id: %[4]d from the %[1]d parcels to %[2]s.t%[3]d_irr table and the junction table records to %[2]s.t%[3]d_jct", yrFrom, nrd, yrTo, parcel.ParcelNo)
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

	rootCmd.AddCommand(jctDeleteCmd)
	jctDeleteCmd.Flags().Int("year", 2020, "Year of JCT to Delete")
	jctDeleteCmd.Flags().Int("parcel_id", 0, "Parcel ID of JCT to Delete")
	jctDeleteCmd.Flags().Int("well_id", 0, "Well ID of JCT to Delete")
	jctDeleteCmd.Flags().String("nrd", "", "NRD of JCT to Delete")

	jctDeleteCmd.MarkFlagRequired("year")
	jctDeleteCmd.MarkFlagRequired("parcel_id")
	jctDeleteCmd.MarkFlagRequired("well_id")
	jctDeleteCmd.MarkFlagRequired("nrd")

	rootCmd.AddCommand(parcelCopyCmd)
	parcelCopyCmd.Flags().Int("year_from", 2020, "Year of Parcel to Copy from")
	parcelCopyCmd.Flags().Int("year_to", 2021, "Year of Parcel to Copy to")
	parcelCopyCmd.Flags().Int("parcel_id", 0, "Parcel ID of Parcel to Copy")
	parcelCopyCmd.Flags().String("nrd", "", "NRD of Parcel to Copy")

	parcelCopyCmd.MarkFlagRequired("year_from")
	parcelCopyCmd.MarkFlagRequired("year_to")
	parcelCopyCmd.MarkFlagRequired("parcel_id")
	parcelCopyCmd.MarkFlagRequired("nrd")
}
