package database

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"
)

type SqliteResult struct {
	Id       int       `db:"id"`
	CellNode int       `db:"cell_node"`
	CellSize float64   `db:"cell_size"`
	Dt       time.Time `db:"dt"`
	FileType int       `db:"file_type"`
	Result   float64   `db:"result"`
}

func (r *SqliteResult) ChangeYear(year int) {
	r.Dt = time.Date(year, r.Dt.Month(), r.Dt.Day(), 0, 0, 0, 0, time.UTC)
}

type SqliteResults []SqliteResult

func GetResults(db *sqlx.DB, year int) (SqliteResults, error) {
	if year < 1953 {
		return nil, errors.New("year must be 1953 or later")
	}

	var results SqliteResults
	query := fmt.Sprintf("SELECT * FROM results WHERE dt >= '%s' AND dt <= '%s' AND file_type NOT IN (124, 116);",
		time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02"),
		time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02"))

	if err := db.Select(&results, query); err != nil {
		return results, err
	}

	return results, nil
}

type SqliteWelResult struct {
	Id       int       `db:"id"`
	WellId   int       `db:"well_id"`
	CellNode int       `db:"cell_node"`
	Dt       time.Time `db:"dt"`
	FileType int       `db:"file_type"`
	Result   float64   `db:"result"`
}

func (r *SqliteWelResult) ChangeYear(year int) {
	r.Dt = time.Date(year, r.Dt.Month(), r.Dt.Day(), r.Dt.Hour(), r.Dt.Minute(), r.Dt.Second(), r.Dt.Nanosecond(), r.Dt.Location())
}

type SqliteWelResults []SqliteWelResult

func GetWelResults(db *sqlx.DB, year int, noRetirements bool) (SqliteWelResults, error) {
	if year < 1953 {
		return nil, errors.New("year must be 1953 or later")
	}

	var results SqliteWelResults
	qryStr := "SELECT * FROM wel_results WHERE dt >= '%s' AND dt <= '%s' AND file_type NOT IN (221, 222);"
	if noRetirements {
		qryStr = "SELECT * FROM wel_results WHERE dt >= '%s' AND dt <= '%s';"
	}

	query := fmt.Sprintf(qryStr,
		time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02"),
		time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02"))

	if err := db.Select(&results, query); err != nil {
		return results, err
	}

	return results, nil
}

func AddRCHRecords(db *sqlx.DB, records SqliteResults) error {
	rchDB, err := ResultsRchDB(db)
	if err != nil {
		return err
	}

	for i := 0; i < len(records); i++ {
		if err := rchDB.Add(RchResult{Node: records[i].CellNode, Size: records[i].CellSize, Dt: records[i].Dt,
			FileType: records[i].FileType, Result: records[i].Result}); err != nil {
			return err
		}
	}

	_ = rchDB.Close()
	return nil
}

func AddWelRecords(db *sqlx.DB, records SqliteWelResults) error {
	rchDB, err := ResultsWelDB(db)
	if err != nil {
		return err
	}

	for i := 0; i < len(records); i++ {
		if err := rchDB.Add(WelResult{
			Wellid:   records[i].WellId,
			Node:     records[i].CellNode,
			Dt:       records[i].Dt,
			FileType: records[i].FileType,
			Result:   records[i].Result,
		}); err != nil {
			return err
		}
	}

	_ = rchDB.Close()
	return nil
}
