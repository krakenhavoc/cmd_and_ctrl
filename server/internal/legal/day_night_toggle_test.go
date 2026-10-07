package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// day_night_toggle_test.go — The Celestus's "{3}, {T}: If it's night,
// it becomes day. Otherwise, it becomes night. Activate only as a
// sorcery." (#2561, ADR 0132), in the agreement style: what the
// enumerator offers the engine accepts, and what is not the sorcery
// window is not offered.

const oracleCelestus = "c0ad2b5f-066b-424b-bddf-d3014731e599"

func TestCelestusToggleIsOfferedInTheSorceryWindowAndAccepted(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	lands(g, active, "Forest", "Forest", 3)
	id := battlefieldCard(g, active, game.Card{
		Name: "The Celestus", TypeLine: "Legendary Artifact", OracleID: oracleCelestus,
	})
	advanceTo(t, g, game.StepPrecombatMain)

	acts := activationsOf(legal.EnumerateFor(g, active.ID), id)
	if len(acts) == 0 {
		t.Fatal("The Celestus's toggle is not offered in the main phase with {3} available")
	}
	dispatchAll(t, g, active.ID, acts)

	// And the offered move really flips it: neither -> night, then
	// night -> day on the clone that resolves it.
	clone := g.Clone()
	m := acts[0]
	if err := actions.Dispatch(clone, actions.Action{
		Type: actions.Type(m.Type), Player: m.Player, Caller: active.ID, Params: m.Params,
	}); err != nil {
		t.Fatalf("dispatch the toggle: %v", err)
	}
	for i := 0; i < 8 && len(clone.StackMeta) > 0; i++ {
		if err := clone.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !clone.IsNight() {
		t.Errorf("the toggle on a game with neither left it %q, want night", clone.DayNightDesignation())
	}
}

func TestCelestusToggleIsWithheldOutsideTheSorceryWindow(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	lands(g, active, "Forest", "Forest", 3)
	id := battlefieldCard(g, active, game.Card{
		Name: "The Celestus", TypeLine: "Legendary Artifact", OracleID: oracleCelestus,
	})
	advanceTo(t, g, game.StepPrecombatMain)
	if len(activationsOf(legal.EnumerateFor(g, active.ID), id)) == 0 {
		t.Fatal("setup: the toggle is not offered in the main phase")
	}
	advanceTo(t, g, game.StepEnd)
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), id); len(acts) != 0 {
		t.Fatalf("the toggle is offered in the end step: %v", labels(acts))
	}
}

func TestCelestusToggleIsWithheldWithoutTheMana(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	lands(g, active, "Forest", "Forest", 2)
	id := battlefieldCard(g, active, game.Card{
		Name: "The Celestus", TypeLine: "Legendary Artifact", OracleID: oracleCelestus,
	})
	advanceTo(t, g, game.StepPrecombatMain)
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), id); len(acts) != 0 {
		t.Fatalf("the toggle is offered with two lands and no other mana: %v", labels(acts))
	}
}
