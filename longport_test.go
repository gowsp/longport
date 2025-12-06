package longport

import (
	"encoding/json"
	"os"
	"testing"
)

func init_longport(t *testing.T) Api {
	os.Setenv("DEBUG_LONGPORT", "1")
	config, err := os.Open("longport_test.json")
	if err != nil {
		t.Fatal(err)
	}
	l := new(Longport)
	err = json.NewDecoder(config).Decode(l)
	if err != nil {
		t.Fatal(err)
	}
	return l
}
