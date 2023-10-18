package rchFiles

import (
	"bitbucket.org/longitude-103/wwumm-esc/database"
	"bitbucket.org/longitude-103/wwumm-esc/parcels"
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"testing"
	"time"
)

var (
	r1 = database.RchResult{Node: 123456, Dt: time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC),
		FileType: 101, Result: 1}

	r2 = database.RchResult{Node: 234567, Dt: time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
		FileType: 101, Result: 2}

	r3 = database.RchResult{Node: 123456, Dt: time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC),
		FileType: 101, Result: 2}

	sliceR = []database.RchResult{r1, r2, r3}

	dp1 = parcels.Parcel{ParcelNo: 101, Nrd: "sp", Ro: [12]float64{0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0}, Dp: [12]float64{0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0}}
	dp2 = parcels.Parcel{ParcelNo: 102, Nrd: "sp", Ro: [12]float64{0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0}, Dp: [12]float64{0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0}}
	dp3 = parcels.Parcel{ParcelNo: 101, Nrd: "np", Ro: [12]float64{0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0}, Dp: [12]float64{0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0}}
	dp4 = parcels.Parcel{ParcelNo: 103, Nrd: "sp", Ro: [12]float64{0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0}, Dp: [12]float64{0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0}}

	sliceDryParcels = []parcels.Parcel{dp1, dp2, dp3, dp4}
)

type AnyTime struct{}

// Match satisfies sqlmock.Argument interface
func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func Test_dryland(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	//sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	rows := sqlmock.NewRows([]string{"node", "c_area", "d_area", "parcel_id", "nrd"}).AddRow(1, 40, 5, 101, "sp").AddRow(2, 40, 5, 102, "sp")

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	mockSqliteDB, mockSqlite, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockSqliteDB.Close()
	sqlitexDB := sqlx.NewDb(mockSqliteDB, "sqlmock")
	mockSqlite.ExpectPrepare("INSERT INTO results")
	rchDB, err1 := database.ResultsRchDB(sqlitexDB)
	if err1 != nil {
		t.Fatalf("an error has occured in rchDB: %s", err1)
	}

	mockSqlite.ExpectBegin()
	mockSqlite.ExpectCommit()
	//mockSqlite.ExpectExec("INSERT INTO results").WithArgs(1, AnyTime{}, 101, 0)

	//v := database.Setup{SYear: 2014, EYear: 2014, RchDb: rchDB, PgDb: sqlxDB}
	//if err2 := Dryland(v, sliceDryParcels); err2 != nil {
	//	t.Errorf("Error in dryland function: %s", err2)
	//}

	if err3 := rchDB.Flush(); err3 != nil {
		t.Errorf("error flushing rchdb: %s", err3)
	}

	// we make sure that all expectations were met
	if err4 := mock.ExpectationsWereMet(); err4 != nil {
		t.Errorf("there were unfulfilled expectations: %s", err4)
	}

	if err5 := mockSqlite.ExpectationsWereMet(); err5 != nil {
		t.Errorf("there were unfulfilled expectations: %s", err5)
	}
}
