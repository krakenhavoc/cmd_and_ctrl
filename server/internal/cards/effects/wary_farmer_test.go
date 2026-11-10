package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const waryFarmerOracle = "63eed1f5-442f-433d-989c-397b88543162"

func TestWaryFarmerIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(waryFarmerOracle)
	if !ok || spec.Name != "Wary Farmer" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// Another creature entered under your control this turn: surveil 1 at
// your end step.
func TestWaryFarmerSurveilsAfterAnotherCreatureEntered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Wary Farmer", "Creature — Kithkin Citizen", waryFarmerOracle, 3, 3)
	advanceToMain(t, g)
	waryFarmerTokenFor(g, me.ID)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) == nil {
		t.Error("no surveil 1 at the end step after another creature entered")
	}
}

// Nothing entered: no trigger at all.
func TestWaryFarmerDoesNothingWhenNoCreatureEntered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	farmer := b12Push(g, me.ID, "Wary Farmer", "Creature — Kithkin Citizen", waryFarmerOracle, 3, 3)
	advanceToMain(t, g)
	advanceToEndStepOf(t, g, 0)
	if triggersOnStackFrom(g, farmer) != 0 || surveilChoiceFor(g, me.ID) != nil {
		t.Error("the Farmer triggered with no creature having entered")
	}
}

// The Farmer entering is not "another" creature; a second creature then
// makes it trigger. An opponent's creature does not count.
func TestWaryFarmerItselfAndOpponentsCreaturesDoNotCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	farmer := rfdCast(t, g, "Wary Farmer", "Creature — Kithkin Citizen", waryFarmerOracle, 3, 3)
	waryFarmerTokenFor(g, opp.ID)
	advanceToEndStepOf(t, g, 0)
	if triggersOnStackFrom(g, farmer) != 0 || surveilChoiceFor(g, me.ID) != nil {
		t.Error("the Farmer counted itself or an opponent's creature")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	rfdCast(t, g2, "Wary Farmer", "Creature — Kithkin Citizen", waryFarmerOracle, 3, 3)
	waryFarmerTokenFor(g2, me2.ID)
	advanceToEndStepOf(t, g2, 0)
	passPriorityAroundTable(t, g2)
	if surveilChoiceFor(g2, me2.ID) == nil {
		t.Error("the Farmer entering plus another creature did not surveil")
	}
}

// waryFarmerTokenFor creates a 1/1 Shapeshifter creature token under
// `who`, an entry the turn's tally records.
func waryFarmerTokenFor(g *game.Game, who uuid.UUID) {
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(who, TokenCard("1/1 colorless Shapeshifter with changeling"), 1)
	})
}
