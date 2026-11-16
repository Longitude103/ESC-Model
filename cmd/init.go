package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes the User Information",
	Long: `Initializes the User Information so that you can log in. Must be run before any other commands.
Host, port, database and SSL mode default to the WWUMM database server, so usually only
the username and password are needed.`,
	Run: func(cmd *cobra.Command, args []string) {
		user, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetString("port")
		dbName, _ := cmd.Flags().GetString("dbname")
		sslMode, _ := cmd.Flags().GetString("sslmode")
		endYear, _ := cmd.Flags().GetInt("endYear")

		newUser := User{
			Username: user,
			Password: password,
			Host:     host,
			Port:     port,
			DBName:   dbName,
			SSLMode:  sslMode,
			EndYear:  endYear,
		}

		// write to file and save in dir
		newUserBytes, err := json.Marshal(newUser)
		if err != nil {
			panic(err)
		}

		// the file holds the password, so keep it private to this user
		err = os.WriteFile(configFile, newUserBytes, 0600)
		if err != nil {
			panic(err)
		}
		if err := os.Chmod(configFile, 0600); err != nil {
			panic(err)
		}

		fmt.Println("Successfully initialized")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	// Here you will define your flags and configuration settings.
	initCmd.Flags().StringP("username", "u", "", "Username for login to database")
	initCmd.Flags().StringP("password", "p", "", "Password for login to database")
	initCmd.Flags().String("host", "", "Database host (default "+defaultHost+")")
	initCmd.Flags().String("port", "", "Database port (default "+defaultPort+")")
	initCmd.Flags().String("dbname", "", "Database name (default "+defaultDBName+")")
	initCmd.Flags().String("sslmode", "", "SSL mode (default "+defaultSSLMode+")")
	initCmd.Flags().Int("endYear", 0, "Last year with model data; the default end year for runModel and runSteadyState averaging (e.g. 2024)")
}
