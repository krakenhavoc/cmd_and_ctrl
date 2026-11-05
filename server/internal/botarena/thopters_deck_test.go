package botarena_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/botarena"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// thopters_deck_test.go is #4991's arena deck: Sai, Master Thopterist at
// the head of a blue artifact deck, each catalog card dealt by a key the
// catalog knows under its name.

func TestThoptersBattleDeck(t *testing.T) {
	deck := botarena.ThoptersBattleDeck([16]byte{})
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
	if !deck[0].IsCommander || deck[0].Name != "Sai, Master Thopterist" {
		t.Errorf("the commander is %q, want Sai, Master Thopterist", deck[0].Name)
	}
	for name, k := range map[string]int{"Hangarback Walker": 4, "Burnished Hart": 4, "Island": 27} {
		if n[name] != k {
			t.Errorf("%d %s, want %d", n[name], name, k)
		}
	}
	if !botarena.IsSynthetic(botarena.ThoptersDeckID) {
		t.Error("thopters-battle needs no Scryfall dump")
	}
}
