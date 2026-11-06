package botarena_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/botarena"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// cant_block_test.go is #4977's arena deck and contestant: BattleDeck
// with Daring Discovery and Untimely Malfunction, each dealt by a catalog
// key the catalog knows, and `heuristic-cantblock`, today's heuristic
// with the price of "can't block" by the attack it opens on.

func TestCantBlockBattleDeck(t *testing.T) {
	deck := botarena.CantBlockBattleDeck([16]byte{})
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
	want := map[string]int{"Daring Discovery": 4, "Untimely Malfunction": 4, "Bear": 6, "Ogre": 6}
	for name, k := range want {
		if n[name] != k {
			t.Errorf("%d %s, want %d", n[name], name, k)
		}
	}
	if !botarena.IsSynthetic(botarena.CantBlockDeckID) {
		t.Error("cantblock-battle needs no Scryfall dump")
	}
}

func TestCantBlockContestant(t *testing.T) {
	const in = "heuristic-cantblock"
	if botarena.CantBlockContestant != in {
		t.Fatalf("CantBlockContestant is %q, want %q", botarena.CantBlockContestant, in)
	}
	got, err := botarena.ParseContestant(in)
	if err != nil {
		t.Fatalf("ParseContestant(%q): %v", in, err)
	}
	if want := (botarena.SeatSpec{Tier: tiers.Heuristic, Variant: botarena.VariantCantBlock}); got != want {
		t.Errorf("ParseContestant(%q) = %+v, want %+v", in, got, want)
	}
	if got.Contestant() != in || got.Label() != in {
		t.Errorf("%q round-trips as contestant %q, label %q", in, got.Contestant(), got.Label())
	}
	if _, err := tiers.Parse(in); err == nil {
		t.Fatalf("tiers.Parse accepts %q; it would be offered in the lobby", in)
	}
}
