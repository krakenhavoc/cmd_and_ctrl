package botarena_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/botarena"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// trigger_mana_test.go is #4770's arena: the `heuristic-triggermana`
// contestant and the ghast-battle deck it is measured on.

func TestTriggerManaContestant(t *testing.T) {
	const in = "heuristic-triggermana"
	if botarena.TriggerManaContestant != in {
		t.Fatalf("TriggerManaContestant is %q, want %q", botarena.TriggerManaContestant, in)
	}
	got, err := botarena.ParseContestant(in)
	if err != nil {
		t.Fatalf("ParseContestant(%q): %v", in, err)
	}
	if want := (botarena.SeatSpec{Tier: tiers.Heuristic, Variant: botarena.VariantTriggerMana}); got != want {
		t.Errorf("ParseContestant(%q) = %+v, want %+v", in, got, want)
	}
	if got.Contestant() != in || got.Label() != in {
		t.Errorf("%q round-trips as contestant %q, label %q", in, got.Contestant(), got.Label())
	}
	if _, err := tiers.Parse(in); err == nil {
		t.Fatalf("tiers.Parse accepts %q; it would be offered in the lobby", in)
	}
}

func TestGhastBattleDeck(t *testing.T) {
	deck := botarena.GhastBattleDeck([16]byte{})
	if len(deck) != 65 {
		t.Errorf("%d cards, want 65", len(deck))
	}
	if !deck[0].IsCommander {
		t.Errorf("the first card %q is not the commander", deck[0].Name)
	}
	n := map[string]int{}
	for _, c := range deck {
		n[c.Name]++
		if c.OracleID == "" {
			continue
		}
		s, ok := effects.Lookup(c.OracleID)
		if !ok || s.Name != c.Name {
			t.Errorf("%s is dealt by %s, which the catalog does not register under that name", c.Name, c.OracleID)
			continue
		}
		if len(s.ManaTriggers) == 0 {
			t.Errorf("%s carries no triggered mana ability", c.Name)
		}
	}
	want := map[string]int{"Swamp": 12, "Forest": 12, "Crypt Ghast": 6, "Zendikar Resurgent": 4}
	for name, k := range want {
		if n[name] != k {
			t.Errorf("%d %s, want %d", n[name], name, k)
		}
	}
	if !botarena.IsSynthetic(botarena.GhastDeckID) {
		t.Error("ghast-battle needs no Scryfall dump")
	}
}
