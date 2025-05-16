package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var user User
var myEnv map[string]string

var rootCmd = &cobra.Command{
	Use:   "EscModel",
	Short: "Esc Model is a CLI tool to interact with the EscModel of the WWUMM",
	Long:  `The EscModel is a CLI tool to interact with the EscModel of the Western Water Use Management Model (WWUMM)`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("EscModel called")
	},
}

func init() {
	cobra.OnInitialize(initConfig)
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func initConfig() {
	// is the json file present?
	if fileInfo, err := os.Stat(".configJson"); os.IsNotExist(err) {
		fmt.Println("config file not found, please run 'EscModel init -h'")
		os.Exit(0)
	} else {
		_ = fileInfo
		//fmt.Println("Using config file:", fileInfo.Name())
		// read the config file
		if configFile, err := os.Open(".configJson"); err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			defer configFile.Close()
			if err := json.NewDecoder(configFile).Decode(&user); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			myEnv = make(map[string]string)
			myEnv["user"] = user.Username
			myEnv["password"] = user.Password
			myEnv["port"] = "5432"
			myEnv["host"] = "pgsql-prd.long103.com"
			myEnv["dbname"] = "wwumm"
		}
	}

	//fmt.Printf("User Info: %+v\n", user)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
