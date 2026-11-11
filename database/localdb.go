package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type lg interface {
	Infof(string, ...interface{})
	Errorf(string, ...interface{})
}

// GetSqlite gets or sets the connection string for the sqlite3 results database.
//
//	It takes no args, and returns the db object for the connection
func GetSqlite(logger lg, path string, fileName string) (*sqlx.DB, error) {
	dbName := fmt.Sprintf("results%s.sqlite", fileName)

	logger.Infof("created sqlite results db named: %s", dbName)
	db, err := sqlx.Open("sqlite3", filepath.FromSlash(filepath.Join(path, dbName)))
	if err != nil {
		logger.Errorf("Error in creating SQLite DB: %s", err)
		return nil, err
	}

	// initializes the database with a results table and file_keys table
	err = InitializeDb(db, logger)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func ConnectSqlite(fileName string) (*sqlx.DB, error) {
	wd, _ := os.Getwd()

	db, err := sqlx.Open("sqlite3", filepath.Join(wd, fileName))
	if err != nil {
		return nil, err
	}

	return db, nil
}
