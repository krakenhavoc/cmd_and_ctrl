package testsu

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

const waryFarmerOracle = "63eed1f5-442f-433d-989c-397b88543162"

func TestWaryFarmerIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(waryFarmerOracle)
	if !ok || spec.Name != "Wary Farmer" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// Another creature entered under your control this turn: surveil 1 at
// your end step.
func TestWaryFarmerSurveilsAfterAnotherCreatureEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Wary Farmer", "Creature — Kithkin Citizen", waryFarmerOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.WaryFarmerTokenFor(g, me.ID)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Error("no surveil 1 at the end step after another creature entered")
	}
}

// Nothing entered: no trigger at all.
func TestWaryFarmerDoesNothingWhenNoCreatureEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	farmer := effectstest.B12Push(g, me.ID, "Wary Farmer", "Creature — Kithkin Citizen", waryFarmerOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	if effectstest.TriggersOnStackFrom(g, farmer) != 0 || effectstest.SurveilChoiceFor(g, me.ID) != nil {
		t.Error("the Farmer triggered with no creature having entered")
	}
}

// The Farmer entering is not "another" creature; a second creature then
// makes it trigger. An opponent's creature does not count.
func TestWaryFarmerItselfAndOpponentsCreaturesDoNotCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	farmer := effectstest.RfdCast(t, g, "Wary Farmer", "Creature — Kithkin Citizen", waryFarmerOracle, 3, 3)
	effectstest.WaryFarmerTokenFor(g, opp.ID)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	if effectstest.TriggersOnStackFrom(g, farmer) != 0 || effectstest.SurveilChoiceFor(g, me.ID) != nil {
		t.Error("the Farmer counted itself or an opponent's creature")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.RfdCast(t, g2, "Wary Farmer", "Creature — Kithkin Citizen", waryFarmerOracle, 3, 3)
	effectstest.WaryFarmerTokenFor(g2, me2.ID)
	effectstest.AdvanceToEndStepOf(t, g2, 0)
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.SurveilChoiceFor(g2, me2.ID) == nil {
		t.Error("the Farmer entering plus another creature did not surveil")
	}
}
