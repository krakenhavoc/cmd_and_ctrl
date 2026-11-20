package effectstest

import (
	"encoding/json"
	"os"
	"testing"
)

// spellText is a catalog card's printed text, from the coverage package's
// oracle fixture; "" when there is none (a face key).
func SpellText(t *testing.T, oracleID string) string {
	t.Helper()
	b, err := os.ReadFile("../../../coverage/testdata/oracle/" + oracleID + ".json")
	if err != nil {
		return ""
	}
	var o struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(b, &o); err != nil {
		t.Fatalf("oracle fixture %s: %v", oracleID, err)
	}
	return o.Text
}
