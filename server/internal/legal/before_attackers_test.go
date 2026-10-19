package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// before_attackers_test.go - the enumerator half of #4062: an ability
// printed "Activate only during your turn, before attackers are declared"
// is a move until the first declare attackers step of the turn begins, and
// every move offered while it is open is one the dispatcher accepts.

const oracleTalasResearcher = "1ed662d7-65b9-4024-a892-c4559fb6aaf9"

func TestBeforeAttackersWindowGatesTheEnumerator(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	talas := battlefieldCard(g, active, game.Card{
		Name: "Talas Researcher", TypeLine: "Creature - Human Pirate Wizard", OracleID: oracleTalasResearcher,
		Power: 1, Toughness: 1,
	})
	advanceTo(t, g, game.StepBeginCombat)

	acts := activationsOf(legal.EnumerateFor(g, active.ID), talas)
	if len(acts) == 0 {
		t.Fatal("in the beginning of combat step the draw is not offered")
	}
	dispatchAll(t, g, active.ID, acts)

	// Attackers declared (the step has begun): the ability is no move.
	advanceTo(t, g, game.StepDeclareAttackers)
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), talas); len(acts) != 0 {
		t.Fatalf("in the declare attackers step the draw is still offered: %v", labels(acts))
	}
}
