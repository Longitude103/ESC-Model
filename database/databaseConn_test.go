package database

import "testing"

func TestPgConnString(t *testing.T) {
	env := map[string]string{
		"user":     "wwumm_office",
		"password": "p@ss:w/rd #1",
		"host":     "db.example.com",
		"port":     "14732",
		"dbname":   "wwumm",
	}

	got := pgConnString(env)
	want := "postgres://wwumm_office:p%40ss%3Aw%2Frd%20%231@db.example.com:14732/wwumm?sslmode=require"
	if got != want {
		t.Errorf("pgConnString() = %q, want %q", got, want)
	}

	env["sslmode"] = "disable"
	got = pgConnString(env)
	want = "postgres://wwumm_office:p%40ss%3Aw%2Frd%20%231@db.example.com:14732/wwumm?sslmode=disable"
	if got != want {
		t.Errorf("pgConnString() with sslmode = %q, want %q", got, want)
	}
}
