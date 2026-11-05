package botarena_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/botarena"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// thopter_deck_test.go is #4991's arena deck: Sai, Master Thopterist and
// her artifacts, each dealt by a catalog key the catalog knows.

func TestThopterBattleDeck(t *testing.T) {
	deck := botarena.ThopterBattleDeck([16]byte{})
	if len(deck) != 65 {
		t.Errorf("%d cards, want 65", len(deck))
	}
	if !deck[0].IsCommander || deck[0].Name != "Sai, Master Thopterist" {
		t.Errorf("the commander is %q, want Sai, Master Thopterist", deck[0].Name)
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
	want := map[string]int{"Island": 24, "Ornithopter": 8, "Memnite": 6, "Hangarback Walker": 4, "Burnished Hart": 4, "Mind Stone": 4}
	for name, k := range want {
		if n[name] != k {
			t.Errorf("%d %s, want %d", n[name], name, k)
		}
	}
	if !botarena.IsSynthetic(botarena.ThopterDeckID) {
		t.Error("thopter-battle needs no Scryfall dump")
	}
}
