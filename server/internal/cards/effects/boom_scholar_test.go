package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// boom_scholar_test.go — #1207 fixed the two root causes Boom
// Scholar shipped as declared caveats (charged-cost display, and the
// CR 601.2f pass reaching a CR 605 mana ability's own cost) but left
// clearing the card itself as a fast follow-up. This file is that
// follow-up's proof: the card's own CostModifier, unchanged since
// #1184, now reaches a MANA ability of another permanent because the
// engine plumbing under it does.
//
// boomScholarManaProbeOracle is a synthetic Signet-shaped exhaust
// mana ability ("{1}, {T}: Add {C}{C}") rather than Loot, the
// Pathfinder's printed "{G}, {T}": a reduction only ever spends
// GENERIC mana (CR 601.2f), and Loot's colored {G} can never show a
// discount landed. mana_ability_cost_modifier_test.go uses the same
// probe shape at the engine layer for the identical reason.
const boomScholarManaProbeOracle = "test-boom-scholar-mana-probe"

func init() {
	Register(Spec{
		OracleID: boomScholarManaProbeOracle,
		Name:     "Boom Scholar Mana Probe",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
			Produced: "{C}{C}",
			Label:    "Exhaust — {1}, {T}: Add {C}{C}.",
			Exhaust:  true,
		}},
	})
}

// TestBoomScholarIsFull: both caveats are gone now that #1207's
// engine fixes are in place.
func TestBoomScholarIsFull(t *testing.T) {
	spec, ok := Lookup(boomScholarOracle)
	if !ok {
		t.Fatal("Boom Scholar is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %+v, want Full and none", spec.Completeness, spec.Caveats)
	}
}

// TestBoomScholarDiscountsAnotherPermanentsExhaustManaAbility closes
// the second caveat ("Abilities that make mana aren't discounted"):
// the probe's exhaust mana ability prints a {1} generic component,
// Boom Scholar's static reduces activation costs by {2}, and the
// floor at zero (CR 601.2f) empties it out completely — proven with
// an EMPTY pool, so paying the printed {1} would fail outright.
func TestBoomScholarDiscountsAnotherPermanentsExhaustManaAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Boom Scholar",
		"Creature — Goblin Advisor", boomScholarOracle, false)
	probe := pushCatalogPermanent(g, me.ID, "Boom Scholar Mana Probe",
		"Artifact", boomScholarManaProbeOracle, false)

	if len(me.ManaPool) != 0 {
		t.Fatalf("pool = %+v, want empty before activating", me.ManaPool)
	}
	if err := g.ActivateManaAbility(me.ID, probe, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("discounted mana ability with an empty pool: %v", err)
	}
	if len(me.ManaPool) != 2 {
		t.Errorf("pool = %+v, want {C}{C} — the {1} generic component is fully discounted", me.ManaPool)
	}
}
