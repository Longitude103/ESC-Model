package database

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestLastExtDataYear(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	v := &Setup{PgDb: sqlx.NewDb(db, "postgres")}

	mock.ExpectQuery(`select max\(yr\) from ext_recharge;`).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(2023))
	yr, err := lastExtDataYear(v, "ext_recharge")
	if err != nil || yr != 2023 {
		t.Errorf("got %d, %v; want 2023", yr, err)
	}

	mock.ExpectQuery(`select max\(yr\) from ext_pumping;`).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(nil))
	if _, err := lastExtDataYear(v, "ext_pumping"); err == nil {
		t.Error("expected an error for an empty table")
	}

	if _, err := lastExtDataYear(v, "parcels; drop table x"); err == nil {
		t.Error("expected an error for an unknown table")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
