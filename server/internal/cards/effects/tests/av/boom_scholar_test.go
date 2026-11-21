package testsav

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestBoomScholarIsFull: both caveats are gone now that #1207's
// engine fixes are in place.
func TestBoomScholarIsFull(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.BoomScholarOracle)
	if !ok {
		t.Fatal("Boom Scholar is not registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Boom Scholar",
		"Creature — Goblin Advisor", effectstest.BoomScholarOracle, false)
	probe := effectstest.PushCatalogPermanent(g, me.ID, "Boom Scholar Mana Probe",
		"Artifact", effectstest.BoomScholarManaProbeOracle, false)

	if me.ManaPool.Total() != 0 {
		t.Fatalf("pool = %+v, want empty before activating", me.ManaPool)
	}
	if err := g.ActivateManaAbility(me.ID, probe, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("discounted mana ability with an empty pool: %v", err)
	}
	if me.ManaPool.Total() != 2 {
		t.Errorf("pool = %+v, want {C}{C} — the {1} generic component is fully discounted", me.ManaPool)
	}
}
