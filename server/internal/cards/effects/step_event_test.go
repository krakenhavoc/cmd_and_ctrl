package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// step_event_test.go — #588: the generic step announcement is a
// trigger event. A probe declares "at the beginning of combat on your
// turn"; it fires on the controller's begin-combat step, goes on the
// stack, and stays quiet on an opponent's.

const stepProbeOracle = "test-step-event-probe"

func init() {
	Register(Spec{
		OracleID: stepProbeOracle,
		Name:     "Step Probe",
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Step Probe — you gain 1 life", Do(GainLife{Amount: 1})),
		},
	})
}

func TestBeginningOfCombatTriggerFiresOnYourTurnOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Step Probe", "Enchantment", stepProbeOracle, false)
	before := me.Life

	// My combat (turn 1, seat 0): on the stack first, then resolved.
	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	if triggerOnStack(g, id) == nil {
		t.Fatal("no trigger on the stack at the beginning of my combat")
	}
	if me.Life != before {
		t.Fatal("the effect ran before the trigger resolved")
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Fatalf("life %d, want %d", me.Life, before+1)
	}

	// The next player's combat: nothing.
	advanceToStepOf(t, g, 1, game.StepBeginCombat)
	if triggerOnStack(g, id) != nil || me.Life != before+1 {
		t.Error("the probe fired on an opponent's beginning of combat")
	}
}

func TestStepBeganPredicateReadsTheTypedStep(t *testing.T) {
	src := &game.Card{Controller: newCatalogGame(t).Seats[0].ID}
	ev := game.Event{Kind: game.EventStepBegan, Step: game.StepEndCombat, Actor: src.Controller, Label: "end_combat"}
	if !StepBegan(game.StepEndCombat, true)(ev, src, game.Characteristic{}, nil) {
		t.Error("StepBegan missed its own step")
	}
	if StepBegan(game.StepBeginCombat, true)(ev, src, game.Characteristic{}, nil) {
		t.Error("StepBegan matched a different step")
	}
	ev.Actor = src.Owner // someone else's turn
	if StepBegan(game.StepEndCombat, true)(ev, src, game.Characteristic{}, nil) {
		t.Error("yours=true matched another player's step")
	}
	if !StepBegan(game.StepEndCombat, false)(ev, src, game.Characteristic{}, nil) {
		t.Error("yours=false should match any player's step")
	}
}
