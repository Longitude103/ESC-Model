package database

import (
	"net"
	"net/url"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PgConnx is a function that returns the sql connection to the postgres database.
func PgConnx(myEnv map[string]string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", pgConnString(myEnv))
	if err != nil {
		return nil, err
	}

	return db, nil
}

// pgConnString builds a postgres:// URL so passwords with special characters
// are escaped. sslmode defaults to require, which the database server enforces.
func pgConnString(myEnv map[string]string) string {
	sslMode := myEnv["sslmode"]
	if sslMode == "" {
		sslMode = "require"
	}

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(myEnv["user"], myEnv["password"]),
		Host:     net.JoinHostPort(myEnv["host"], myEnv["port"]),
		Path:     "/" + myEnv["dbname"],
		RawQuery: url.Values{"sslmode": {sslMode}}.Encode(),
	}
	return u.String()
}
