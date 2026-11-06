package botarena_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/botarena"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// xpumps_test.go is #4976's arena deck and contestant: BattleDeck in red
// and green with the pumps of the announced X in it, each dealt by a
// catalog key the catalog knows, and `heuristic-xpumps`, today's
// heuristic with their price on.

func TestXPumpsBattleDeck(t *testing.T) {
	deck := botarena.XPumpsBattleDeck([16]byte{})
	if len(deck) != len(botarena.BattleDeck([16]byte{})) {
		t.Errorf("%d cards, want BattleDeck's %d", len(deck), len(botarena.BattleDeck([16]byte{})))
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
		if c.Name == "Lightning Bolt" {
			continue
		}
		declared := false
		for _, p := range append([]effects.ActivatedAbility{{Purpose: s.Purpose}}, s.Activated...) {
			for _, e := range p.Purpose.Targets.List() {
				declared = declared || e.PumpPowerIsX
			}
		}
		if s.Modes != nil {
			for _, o := range s.Modes.Options {
				for _, e := range o.Purpose.Targets.List() {
					declared = declared || e.PumpPowerIsX
				}
			}
		}
		if !declared {
			t.Errorf("%s declares no pump of X for the arena to measure", c.Name)
		}
	}
	want := map[string]int{"Tyvar's Stand": 4, "Mishra's Command": 4, "Primal Might": 2, "Kessig Wolf Run": 2, "Forest": 10, "Mountain": 14}
	for name, k := range want {
		if n[name] != k {
			t.Errorf("%d %s, want %d", n[name], name, k)
		}
	}
	if !botarena.IsSynthetic(botarena.XPumpsDeckID) {
		t.Error("xpumps-battle needs no Scryfall dump")
	}
}

func TestXPumpsContestant(t *testing.T) {
	const in = "heuristic-xpumps"
	if botarena.XPumpsContestant != in {
		t.Fatalf("XPumpsContestant is %q, want %q", botarena.XPumpsContestant, in)
	}
	got, err := botarena.ParseContestant(in)
	if err != nil {
		t.Fatalf("ParseContestant(%q): %v", in, err)
	}
	if want := (botarena.SeatSpec{Tier: tiers.Heuristic, Variant: botarena.VariantXPumps}); got != want {
		t.Errorf("ParseContestant(%q) = %+v, want %+v", in, got, want)
	}
	if _, err := tiers.Parse(in); err == nil {
		t.Fatalf("tiers.Parse accepts %q; it would be offered in the lobby", in)
	}
}
