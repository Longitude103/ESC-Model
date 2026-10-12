package cmd

import (
	"os"
	"testing"
)

// A config file from before EM-64 only has username/password; it should
// connect to the default server with SSL.
func TestInitConfigDefaults(t *testing.T) {
	dir := t.TempDir()
	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configFile, []byte(`{"username":"u","password":"p"}`), 0600); err != nil {
		t.Fatal(err)
	}

	initConfig()

	want := map[string]string{
		"user": "u", "password": "p",
		"host": defaultHost, "port": defaultPort, "dbname": defaultDBName, "sslmode": defaultSSLMode,
	}
	for k, v := range want {
		if myEnv[k] != v {
			t.Errorf("myEnv[%q] = %q, want %q", k, myEnv[k], v)
		}
	}
}
