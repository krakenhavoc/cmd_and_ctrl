package testso

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// optional_life_cards_test.go — ADR 0100's 2026-11-20 amendment
// (#5782): a kicker or buyback that pays a fixed amount of life.
//
//   - Slaughter: "Buyback—Pay 4 life", paid with the total cost (CR
//     601.2h) and refused below 4 life (CR 119.4).
//   - Phyrexian Scuta: "Kicker—Pay 3 life", read back as a kick.

const (
	slaughterOracle      = "37de575d-15d4-42a2-a3d2-3855fb6ab110"
	phyrexianScutaOracle = "af2845cf-8770-4efa-8a83-bf9eb4122b78"
)

func TestOptionalLifeCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		slaughterOracle:      "Slaughter",
		phyrexianScutaOracle: "Phyrexian Scuta",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
		if len(spec.OptionalCosts) != 1 || spec.OptionalCosts[0].PayLife == 0 {
			t.Errorf("%s: optional costs %+v, want one that pays life", name, spec.OptionalCosts)
		}
	}
}

func TestSlaughterBuybackPaysFourLife(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		wantHand bool
		wantLife int
	}{
		{"not bought back goes to the graveyard and pays no life", nil, false, 40},
		{"bought back pays 4 life and returns to hand", []int{0}, true, 36},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			caster := g.Seats[g.Turn.ActiveSeat]
			caster.Life = 40
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			bear := effectstest.SeedCreature(g, "Grizzly Bears", opp.ID)

			id, err := effectstest.CastWithOptionalCosts(t, g, "Slaughter", "Instant", slaughterOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}, tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			// CR 601.2h: the life is paid as the spell is cast, before it
			// resolves.
			if caster.Life != tc.wantLife {
				t.Errorf("life after the cast = %d, want %d", caster.Life, tc.wantLife)
			}
			effectstest.PassPriorityAroundTable(t, g)

			if effectstest.FindBattlefieldCardByID(g, bear) != nil {
				t.Errorf("the nonblack creature survived Slaughter")
			}
			inHand, inGrave := caster.Hand.Contains(id), caster.Graveyard.Contains(id)
			if inHand != tc.wantHand || inGrave == tc.wantHand {
				t.Errorf("Slaughter after resolution: hand=%v graveyard=%v, want hand=%v", inHand, inGrave, tc.wantHand)
			}
		})
	}
}

// CR 119.4: a caster below 4 life can't announce the buyback, and the
// refusal leaves the card in hand and the life total alone.
func TestSlaughterRefusesTheBuybackBelowFourLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	caster.Life = 3
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.SeedCreature(g, "Grizzly Bears", opp.ID)

	id, err := effectstest.CastWithOptionalCosts(t, g, "Slaughter", "Instant", slaughterOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}, []int{0}, nil)
	if err == nil {
		t.Fatal("a buyback paying 4 life was accepted at 3 life")
	}
	if !caster.Hand.Contains(id) {
		t.Error("a refused cast did not leave the card in hand")
	}
	if caster.Life != 3 {
		t.Errorf("life after a refused cast = %d, want 3", caster.Life)
	}
}

func TestPhyrexianScutaKickerPaysThreeLifeForTwoCounters(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		counters int
		wantLife int
	}{
		{"unkicked", nil, 0, 40},
		{"kicked", []int{0}, 2, 37},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			caster := g.Seats[g.Turn.ActiveSeat]
			caster.Life = 40
			id, err := effectstest.CastWithOptionalCosts(t, g, "Phyrexian Scuta", "Creature — Phyrexian Zombie", phyrexianScutaOracle, nil, tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.PlusOneCounters(g, id); n != tc.counters {
				t.Errorf("+1/+1 counters %d, want %d", n, tc.counters)
			}
			if caster.Life != tc.wantLife {
				t.Errorf("life %d, want %d", caster.Life, tc.wantLife)
			}
		})
	}
}
