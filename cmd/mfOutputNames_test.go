package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOutputName(t *testing.T) {
	cases := []struct{ base, suffix, want string }{
		{"AggregateWEL", "", "AggregateWEL"},
		{"AggregateWEL", "npnrd", "AggregateWEL_npnrd"},
		{"AggregateRCH", " task3 npnrd ", "AggregateRCH_task3_npnrd"},
		{"201 - NP Comingled Pre 1998WEL", "_npnrd", "201_-_NP_Comingled_Pre_1998WEL_npnrd"},
		{"AggregateWEL", "a/b\\c", "AggregateWEL_a_b_c"},
	}
	for _, c := range cases {
		if got := outputName(c.base, c.suffix); got != c.want {
			t.Errorf("outputName(%q, %q) = %q, want %q", c.base, c.suffix, got, c.want)
		}
	}
}

const welSample = `# MODFLOW6 Well Boundary Package
# Desc: base run
BEGIN OPTIONS
  SAVE_FLOWS
  AUTO_FLOW_REDUCE 1.000000e-01
  AUTO_FLOW_REDUCE_CSV FILEOUT Desc:_base_run.csv
END OPTIONS

BEGIN PERIOD 1
 12 -1.000000e+02
END PERIOD 1
`

func TestRewriteReduceCSV(t *testing.T) {
	var out bytes.Buffer
	if err := rewriteReduceCSV(strings.NewReader(welSample), &out, "AggregateWEL_npnrd_reduction.csv"); err != nil {
		t.Fatal(err)
	}

	want := strings.Replace(welSample, "FILEOUT Desc:_base_run.csv", "FILEOUT AggregateWEL_npnrd_reduction.csv", 1)
	if out.String() != want {
		t.Errorf("unexpected output:\n%s", out.String())
	}
}

func TestSetReduceCSVName(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "AggregateWEL_npnrd.WEL6")
	if err := os.WriteFile(p, []byte(welSample), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := setReduceCSVName(p, "AggregateWEL_npnrd_reduction.csv"); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "AUTO_FLOW_REDUCE_CSV FILEOUT AggregateWEL_npnrd_reduction.csv\n") {
		t.Errorf("CSV name not rewritten:\n%s", b)
	}
	if !strings.Contains(string(b), "BEGIN PERIOD 1\n 12 -1.000000e+02\nEND PERIOD 1\n") {
		t.Errorf("body changed:\n%s", b)
	}
	if _, err := os.Stat(p + ".tmp"); !os.IsNotExist(err) {
		t.Error("temp file left behind")
	}
}
