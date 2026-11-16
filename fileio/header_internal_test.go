package fileio

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Longitude103/ESC-Model/logging"
)

func dataLine(yr string) string {
	f := []string{"AGAT", yr, "412", "10", "1", "3"}
	for i := 0; i < 72; i++ {
		f = append(f, "0.10")
	}
	return strings.Join(f, " ")
}

func TestGetFileData_FirstLine(t *testing.T) {
	dir := t.TempDir()
	logger := logging.NewLogger(filepath.Join(dir, "log"))
	defer logger.Close()

	noHeader := filepath.Join(dir, "AGAT_MON.TXT")
	_ = os.WriteFile(noHeader, []byte(dataLine("1953")+"\n"+dataLine("1954")+"\n"), 0o644)
	rows, err := getFileData(noHeader, logger)
	if err != nil || len(rows) != 2 || rows[0].Yr != 1953 {
		t.Fatalf("file without header: got %d rows (first year %v), err %v; want 2 rows starting 1953", len(rows), rows, err)
	}

	withHeader := filepath.Join(dir, "AKRN_MON.TXT")
	_ = os.WriteFile(withHeader, []byte("STN YEAR SOIL CROP TILL IRR ...\n"+dataLine("1953")+"\n"), 0o644)
	rows, err = getFileData(withHeader, logger)
	if err != nil || len(rows) != 1 || rows[0].Yr != 1953 {
		t.Fatalf("file with header: got %d rows, err %v; want 1 row for 1953", len(rows), err)
	}
}
