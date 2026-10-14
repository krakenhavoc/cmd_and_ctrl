package cards

import (
	"os"
	"testing"

	"github.com/google/uuid"
)

// Manual check against the real Scryfall dump (#3605): the card the
// issue was filed about renders its Streets of New Capenna art. Skips
// unless CMDCTRL_SCRYFALL_DUMP points at default-cards.json.
func TestRealDumpRaffineDefaultsToSNC(t *testing.T) {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("CMDCTRL_SCRYFALL_DUMP not set")
	}
	idx := NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatal(err)
	}
	c, ok := idx.FindByName("Raffine, Scheming Seer")
	if !ok {
		t.Fatal("Raffine not in the dump")
	}
	if c.SetCode != "snc" || c.CollectorNo != "213" {
		t.Errorf("default printing is %s #%s (%s), want snc #213", c.SetCode, c.CollectorNo, c.SetType)
	}
	byOracle, _ := idx.FindByOracleID(c.OracleID)
	if byOracle.ID != c.ID {
		t.Errorf("byOracle and byName disagree: %s vs %s", byOracle.ID, c.ID)
	}
	p, ok := idx.FindPrintingOf(c, c.ID, "", "")
	if !ok || p.ID != c.ID {
		t.Errorf("the default printing does not resolve by id")
	}
	if p, ok := idx.FindPrintingOf(c, uuid.Nil, "SLZ", "93"); !ok || p.SetCode != "slz" {
		t.Errorf("slz #93 by set and number: %v %s", ok, p.SetCode)
	}
}
