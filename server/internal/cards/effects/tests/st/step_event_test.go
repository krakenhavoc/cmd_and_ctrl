package testsst

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestBeginningOfCombatTriggerFiresOnYourTurnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Step Probe", "Enchantment", effectstest.StepProbeOracle, false)
	before := me.Life

	// My combat (turn 1, seat 0): on the stack first, then resolved.
	effectstest.AdvanceToStepOf(t, g, 0, game.StepBeginCombat)
	if effectstest.TriggerOnStack(g, id) == nil {
		t.Fatal("no trigger on the stack at the beginning of my combat")
	}
	if me.Life != before {
		t.Fatal("the effect ran before the trigger resolved")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Fatalf("life %d, want %d", me.Life, before+1)
	}

	// The next player's combat: nothing.
	effectstest.AdvanceToStepOf(t, g, 1, game.StepBeginCombat)
	if effectstest.TriggerOnStack(g, id) != nil || me.Life != before+1 {
		t.Error("the probe fired on an opponent's beginning of combat")
	}
}

func TestStepBeganPredicateReadsTheTypedStep(t *testing.T) {
	src := &game.Card{Controller: effectstest.NewCatalogGame(t).Seats[0].ID}
	ev := game.Event{Kind: game.EventStepBegan, Step: game.StepEndCombat, Actor: src.Controller, Label: "end_combat"}
	if !effects.StepBegan(game.StepEndCombat, true)(ev, src, game.Characteristic{}, nil) {
		t.Error("StepBegan missed its own step")
	}
	if effects.StepBegan(game.StepBeginCombat, true)(ev, src, game.Characteristic{}, nil) {
		t.Error("StepBegan matched a different step")
	}
	ev.Actor = src.Owner // someone else's turn
	if effects.StepBegan(game.StepEndCombat, true)(ev, src, game.Characteristic{}, nil) {
		t.Error("yours=true matched another player's step")
	}
	if !effects.StepBegan(game.StepEndCombat, false)(ev, src, game.Characteristic{}, nil) {
		t.Error("yours=false should match any player's step")
	}
}
