package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestMayMakeManaUnautomatedOnTheWire: an uncatalogued card that prints
// a mana ability is flagged; one without is not, and the key is
// omitted when false (#2258, ADR 0118 decision 8).
func TestMayMakeManaUnautomatedOnTheWire(t *testing.T) {
	prev := game.IsCatalogCard
	game.IsCatalogCard = func(string) bool { return false }
	t.Cleanup(func() { game.IsCatalogCard = prev })

	rock := game.Card{InstanceID: uuid.New(), Name: "Rock", OracleID: "r", NeedsEffect: true, PrintsManaAbility: true}
	if v := viewOfCard(rock); !v.MayMakeManaUnautomated {
		t.Error("uncatalogued mana rock: want may_make_mana_unautomated")
	}
	plain := game.Card{InstanceID: uuid.New(), Name: "Plain", OracleID: "p", NeedsEffect: true}
	v := viewOfCard(plain)
	if v.MayMakeManaUnautomated {
		t.Error("uncatalogued non-mana card: flag set")
	}
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "may_make_mana_unautomated") {
		t.Error("false flag should be omitted from the wire")
	}

	game.IsCatalogCard = func(string) bool { return true }
	if v := viewOfCard(rock); v.MayMakeManaUnautomated {
		t.Error("catalogued mana rock: flag set")
	}
}
