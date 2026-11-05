package botarena_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/botarena"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// haste_deck_test.go is #4981's arena deck: BattleDeck with haste grants
// and creatures whose value is a {T} ability, each dealt by a catalog key
// the catalog knows, so #4324's PriceHasteTapRows has something to price.

func TestHasteBattleDeck(t *testing.T) {
	deck := botarena.HasteBattleDeck([16]byte{})
	if len(deck) != len(botarena.BattleDeck([16]byte{})) {
		t.Errorf("%d cards, want BattleDeck's %d", len(deck), len(botarena.BattleDeck([16]byte{})))
	}
	n := map[string]int{}
	lands := 0
	for _, c := range deck {
		n[c.Name]++
		if c.TypeLine == "Land" || c.TypeLine == "Basic Land — Mountain" {
			lands++
		}
		if c.OracleID == "" {
			continue
		}
		if s, ok := effects.Lookup(c.OracleID); !ok || s.Name != c.Name {
			t.Errorf("%s is dealt by %s, which the catalog does not register under that name", c.Name, c.OracleID)
		}
	}
	want := map[string]int{"Expedite": 6, "Hanweir Battlements": 2, "Frostwielder": 4, "Fervent Paincaster": 4, "Goblin Sharpshooter": 3, "Mountain": 22}
	for name, k := range want {
		if n[name] != k {
			t.Errorf("%d %s, want %d", n[name], name, k)
		}
	}
	if lands != 24 {
		t.Errorf("%d lands, want BattleDeck's 24", lands)
	}
	if !botarena.IsSynthetic(botarena.HasteDeckID) {
		t.Error("haste-battle needs no Scryfall dump")
	}
}
