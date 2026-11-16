package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// configFile holds the user's database login, written by `EscModel init`.
const configFile = ".configjson"

// Defaults for the WWUMM database on the Linode managed PostgreSQL server.
// A config file can override any of them.
const (
	defaultHost    = "a538082-akamai-prod-2405558-default.g2a.akamaidb.net"
	defaultPort    = "14732"
	defaultDBName  = "wwumm"
	defaultSSLMode = "require"
)

// fallbackEndYear is only used when the config file has no endYear (written by older versions of init).
const fallbackEndYear = 2020

// configEndYear is the last year with model data, read from the config file. runModel and runSteadyState use it as
// their default end year, so extending the model only needs `init --endYear`.
var configEndYear int

var user User
var myEnv map[string]string

var rootCmd = &cobra.Command{
	Use:   "EscModel",
	Short: "Esc Model is a CLI tool to interact with the EscModel of the WWUMM",
	Long:  `The EscModel is a CLI tool to interact with the EscModel of the Western Water Use Management Model (WWUMM)`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// init creates the config file, so it can't require one
		switch cmd.Name() {
		case "init", "help", "completion", "EscModel":
			return
		}
		initConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("EscModel called")
	},
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host,omitempty"`
	Port     string `json:"port,omitempty"`
	DBName   string `json:"dbname,omitempty"`
	SSLMode  string `json:"sslmode,omitempty"`
	EndYear  int    `json:"endYear,omitempty"`
}

func initConfig() {
	// is the json file present?
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("config file not found, please run 'EscModel init -h'")
		os.Exit(0)
	}

	f, err := os.Open(configFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&user); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	myEnv = make(map[string]string)
	myEnv["user"] = user.Username
	myEnv["password"] = user.Password
	myEnv["host"] = valueOr(user.Host, defaultHost)
	myEnv["port"] = valueOr(user.Port, defaultPort)
	myEnv["dbname"] = valueOr(user.DBName, defaultDBName)
	myEnv["sslmode"] = valueOr(user.SSLMode, defaultSSLMode)

	configEndYear = user.EndYear
	if configEndYear == 0 {
		configEndYear = fallbackEndYear
		fmt.Printf("No endYear in %s, using %d. Run 'EscModel init --endYear <year>' to set it.\n", configFile, fallbackEndYear)
	}
}

func valueOr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
