package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestCardStrippedByARestoreShowsManual is #522's wire half: a card a
// restore brought back with fewer catalog abilities than were captured
// is shown as a manual card — `unimplemented`, which is the client's
// `manual` chip, and never `auto` — even while its (shrunken) catalog
// entry still exists.
func TestCardStrippedByARestoreShowsManual(t *testing.T) {
	const oracle = "00000000-0000-4000-8000-00000000522b"
	prev := game.CatalogLookup
	t.Cleanup(func() { game.CatalogLookup = prev })
	game.CatalogLookup = func(k string) *game.CardDef {
		if k == oracle {
			return &game.CardDef{}
		}
		return nil
	}

	c := game.Card{InstanceID: uuid.New(), Name: "Parity Probe", OracleID: oracle}
	if v := viewOfCard(c); !v.Auto || v.Unimplemented {
		t.Fatalf("setup: an intact catalog card reads auto=%v unimplemented=%v", v.Auto, v.Unimplemented)
	}
	c.AbilitiesLostOnRestore = true
	if v := viewOfCard(c); v.Auto || !v.Unimplemented {
		t.Errorf("a stripped card reads auto=%v unimplemented=%v, want false/true", v.Auto, v.Unimplemented)
	}
}
