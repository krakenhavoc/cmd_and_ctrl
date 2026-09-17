package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// activation_condition_test.go — the enumerator half of #743. An
// activated ability whose "Activate only if …" condition is false is
// not a move (the engine refuses it with ErrConditionNotMet), and the
// moment the condition holds every offered activation is one the
// dispatcher accepts (#544).

const oracleTectonicEdge = "4927150d-7ff6-4232-b20e-d2ea245ac710"

func TestActivationConditionGatesTheEnumerator(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	edge := battlefieldCard(g, active, game.Card{Name: "Tectonic Edge", TypeLine: "Land", OracleID: oracleTectonicEdge})
	battlefieldCard(g, opp, game.Card{Name: "Cabal Coffers", TypeLine: "Land"})
	battlefieldCard(g, opp, basic("Swamp", "Swamp"))
	battlefieldCard(g, opp, basic("Swamp", "Swamp"))
	mana(g, active, 1)
	advanceTo(t, g, game.StepPrecombatMain)

	// Three lands: the destroy is not a move. The mana ability still is.
	moves := legal.EnumerateFor(g, active.ID)
	if acts := activationsOf(moves, edge); len(acts) != 0 {
		t.Fatalf("no opponent on four lands: want no activation, got %v", labels(acts))
	}
	if !hasManaMoveFrom(moves, edge) {
		t.Error("the ungated {T}: Add {C} should still be offered")
	}

	// A fourth land opens it, and every offer is accepted.
	battlefieldCard(g, opp, basic("Swamp", "Swamp"))
	moves = legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, edge)
	if len(acts) == 0 {
		t.Fatal("an opponent on four lands: want the destroy offered")
	}
	dispatchAll(t, g, active.ID, acts)
}

func hasManaMoveFrom(moves []legal.Move, source interface{ String() string }) bool {
	for _, m := range moves {
		if m.Kind == legal.KindMana && m.Source.String() == source.String() {
			return true
		}
	}
	return false
}
