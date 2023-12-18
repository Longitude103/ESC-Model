package database

import (
	"bitbucket.org/longitude-103/wwumm-esc/Utils"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type FileKeys struct {
	FileKey     int    `db:"file_key"`
	Description string `db:"description"`
}

func (f FileKeys) Print() string {
	return fmt.Sprintf("%d - %s", f.FileKey, f.Description)
}

type MfResults struct {
	CellNode       int             `db:"cell_node"`
	CellSize       sql.NullFloat64 `db:"cell_size"`
	ResultDate     time.Time       `db:"dt"`
	Rslt           float64         `db:"rslt"`
	Rw             sql.NullInt64   `db:"rw"`
	Clm            sql.NullInt64   `db:"clm"`
	ConvertedValue bool            // value in Rslt is already converted, no need to use convert methods
}

type ResultsNote struct {
	Id   int    `db:"id"`
	Note string `db:"note"`
}

func (m MfResults) Date() time.Time {
	return m.ResultDate
}

func (m MfResults) Node() int {
	return m.CellNode
}

func (m MfResults) Value() float64 {
	return m.Rslt
}

func (m MfResults) RowCol() (int, int) {
	return int(m.Rw.Int64), int(m.Clm.Int64)
}

func (m MfResults) IsNodeResult() bool {
	if m.Rw.Valid && m.Clm.Valid {
		return false
	}

	return true
}

func (m MfResults) Year() int {
	return m.ResultDate.Year()
}

func (m MfResults) Month() int {
	return int(m.ResultDate.Month())
}

func (m MfResults) ConvertToFtPDay() float64 {
	return (m.Rslt / m.CellSize.Float64) / float64(Utils.TimeExt{T: m.ResultDate}.DaysInMonth())
}

func (m MfResults) ConvertToFt3PDay() float64 {
	return (m.Rslt * 43560) / float64(Utils.TimeExt{T: m.ResultDate}.DaysInMonth()) * -1
}

func (m MfResults) UseValue() bool {
	return m.ConvertedValue
}

func (m *MfResults) SetConvertedValue() {
	m.ConvertedValue = true
}

func GetFileKeys(db *sqlx.DB, wel bool) ([]string, error) {
	var fKeys []FileKeys
	var resultFileKeys []string
	var query string
	if wel { // give the wel file_keys
		query = "SELECT file_key, description FROM file_keys WHERE file_key > 199;"
	} else { // give the rch file_keys
		query = "SELECT file_key, description FROM file_keys WHERE file_key < 200;"
	}

	if err := db.Select(&fKeys, query); err != nil {
		return resultFileKeys, err
	}

	for _, k := range fKeys {
		resultFileKeys = append(resultFileKeys, k.Print())
	}

	return resultFileKeys, nil
}

var (
	//go:embed sql/AggResultsNoExclude.sql
	AggResultsNoExclude string
	//go:embed sql/RchAggResultsNoExclude.sql
	RchAggResultsNoExclude string
	//go:embed sql/WelSingleResult.sql
	WelSingleResult string
	//go:embed sql/RchSingleResult.sql
	RchSingleResult string
)

func GetAggResults(db *sqlx.DB, wel bool, excludeList []string) ([]MfResults, error) {
	var qry string
	var results []MfResults

	// TODO: Write a check here to ensure that cellrc table has records

	if wel { // is a wel file
		if len(excludeList) > 0 { // has an item in the to exclude list
			list := excludeList[0][0:3]
			for i := 1; i < len(excludeList); i++ {
				list += ", "
				list += excludeList[i][0:3]
			}

			qry = fmt.Sprintf(`SELECT cell_node, rw, clm, dt, rslt
									from (SELECT cell_node, rw, clm, dt, sum(result) rslt
 								  FROM wel_results LEFT JOIN cellrc on cell_node = node WHERE file_type NOT IN (%s)
									group by cell_node, dt) where rslt > 0;`, list)
		} else { // don't exclude anything
			qry = AggResultsNoExclude
		}
	} else { // is a recharge file
		if len(excludeList) > 0 { // has an item in exclude list
			list := excludeList[0][0:3]
			for i := 1; i < len(excludeList); i++ {
				list += ", "
				list += excludeList[i][0:3]
			}

			qry = fmt.Sprintf(`SELECT cell_node, cell_size, rw, clm, dt, rslt from (SELECT cell_node, cell_size, rw, 
									clm, dt, sum(result) rslt FROM results LEFT JOIN cellrc on cell_node = node 
								    WHERE file_type NOT IN (%s) group by cell_node, cell_size, dt) where rslt > 0;`, list)
		} else { // don't exclude anything
			qry = RchAggResultsNoExclude
		}
	}

	if err := db.Select(&results, qry); err != nil {
		return results, err
	}

	return results, nil
}

func SingleResult(db *sqlx.DB, wel bool, fileKey string) ([]MfResults, error) {
	var results []MfResults
	var qry string

	if wel {
		qry = WelSingleResult
	} else {
		qry = RchSingleResult
	}

	if err := db.Select(&results, qry, fileKey[0:3]); err != nil {
		return results, err
	}

	return results, nil
}

func GetDescription(db *sqlx.DB) (desc string, err error) {
	rslt, err := getAllDBResults(db)
	if err != nil {
		return "", err
	}

	for _, n := range rslt {
		if strings.ToLower(n.Note[:4]) == "desc" {
			return n.Note, nil
		}
	}

	return "", errors.New("could not find description")
}

func GetGrid(db *sqlx.DB) (grid int, err error) {
	rslt, err := getAllDBResults(db)
	if err != nil {
		return 0, err
	}

	for _, n := range rslt {
		if strings.ToLower(n.Note[:4]) == "grid" {
			i, _ := strconv.Atoi(n.Note[len(n.Note)-1:])
			return i, nil
		}
	}

	return 0, errors.New("could not find grid number")
}

func GetStartEndYrs(db *sqlx.DB) (SYr, EYr int, err error) {
	rslt, err := getAllDBResults(db)
	if err != nil {
		return 0, 0, err
	}

	for _, n := range rslt {
		if strings.ToLower(n.Note[:3]) == "sta" {
			i, _ := strconv.Atoi(n.Note[len(n.Note)-4:])
			SYr = i
		}

		if strings.ToLower(n.Note[:3]) == "end" {
			i, _ := strconv.Atoi(n.Note[len(n.Note)-4:])
			EYr = i
		}
	}

	for _, n := range rslt {
		if len(n.Note) > 21 {
			if strings.ToLower(n.Note[:22]) == "robust review end year" {
				i, _ := strconv.Atoi(n.Note[len(n.Note)-4:])
				EYr = i
			}
		}
	}

	if SYr == 0 || EYr == 0 {
		return 0, 0, errors.New("could not find start/end year")
	}

	return SYr, EYr, nil
}

func GetSteadyState(db *sqlx.DB) (bool, error) {
	rslt, err := getAllDBResults(db)
	if err != nil {
		return false, err
	}

	for _, n := range rslt {
		if strings.ToLower(n.Note[:6]) == "steady" {
			return true, nil
		}
	}

	return false, nil
}

func getAllDBResults(db *sqlx.DB) ([]ResultsNote, error) {
	var rslt []ResultsNote
	query := "SELECT * FROM results_notes"

	if err := db.Select(&rslt, query); err != nil {
		return rslt, err
	}

	return rslt, nil
}
