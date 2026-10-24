package botarena_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/botarena"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// tricks_deck_test.go is #4312's arena deck: BattleDeck with the changes
// until end of turn the heuristic prices in it, each dealt by a catalog
// key the catalog knows.

func TestTricksBattleDeck(t *testing.T) {
	deck := botarena.TricksBattleDeck([16]byte{})
	if len(deck) != len(botarena.BattleDeck([16]byte{})) {
		t.Errorf("%d cards, want BattleDeck's %d", len(deck), len(botarena.BattleDeck([16]byte{})))
	}
	n := map[string]int{}
	for _, c := range deck {
		n[c.Name]++
		if c.OracleID == "" {
			continue
		}
		if s, ok := effects.Lookup(c.OracleID); !ok || s.Name != c.Name {
			t.Errorf("%s is dealt by %s, which the catalog does not register under that name", c.Name, c.OracleID)
		}
	}
	want := map[string]int{"Brute Force": 4, "Dismember": 3, "Mother of Runes": 4, "Center Soul": 3, "Plains": 7, "Rogue's Passage": 1}
	for name, k := range want {
		if n[name] != k {
			t.Errorf("%d %s, want %d", n[name], name, k)
		}
	}
	if !botarena.IsSynthetic(botarena.TricksDeckID) {
		t.Error("tricks-battle needs no Scryfall dump")
	}
}
