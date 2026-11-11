package database

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type fakelog struct {
	m string
}

func (l *fakelog) Errorf(template string, args ...interface{}) {
	l.m = fmt.Sprintf(template, args...)
}

func (l *fakelog) Infof(template string, args ...interface{}) {
	l.m = fmt.Sprintf(template, args...)
}

func TestGetSqlite(t *testing.T) {
	l := &fakelog{}

	dir := t.TempDir()
	db, err := GetSqlite(l, dir, "testfile")
	if err != nil {
		t.Fatalf("Error Creating SQLite: %s", err)
	}
	defer db.Close()

	if _, err := os.Stat(filepath.Join(dir, "resultstestfile.sqlite")); os.IsNotExist(err) {
		t.Error("Didn't create the SQLite Database")
	}

}
