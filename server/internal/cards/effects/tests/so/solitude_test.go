package testsso

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Solitude's enters trigger exiles up to ONE OTHER target creature and
// its controller gains life equal to its power (#2522).
func TestSolitudeExilesAnotherCreatureAndGivesItsControllerLifeEqualToPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Brute", "Creature — Ogre", 4, 4)
	life := opp.Life

	sol := effectstest.CastAndResolveCreature(t, g, "Solitude", "Creature — Elemental Incarnation", effectstest.SolitudeOracle)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for Solitude's enters trigger")
	}
	if effectstest.HasID(pick.PickTargetCards, sol) {
		t.Error("Solitude is offered as its own target; the clause says OTHER")
	}
	if !effectstest.HasID(pick.PickTargetCards, victim) {
		t.Fatalf("the opponent's creature is not a legal target: %v", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.B12ZoneOf(g, victim) != game.ZoneExile {
		t.Error("the target is exiled")
	}
	if opp.Life != life+4 {
		t.Errorf("the creature's controller gains its power: life %d -> %d, want +4", life, opp.Life)
	}
	if !g.Battlefield.Contains(sol) {
		t.Error("a hard-cast Solitude stays")
	}
}

// "Up to one": with no other creature the trigger is removed and
// Solitude stays; a target is never forced.
func TestSolitudeWithNoOtherCreatureAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sol := effectstest.CastAndResolveCreature(t, g, "Solitude", "Creature — Elemental Incarnation", effectstest.SolitudeOracle)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a target prompt opened with no other creature on the battlefield")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sol) {
		t.Error("Solitude exiled itself")
	}
}
