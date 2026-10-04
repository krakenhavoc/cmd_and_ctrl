package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// ADR 0117 §5, the enumerator half: a mana ability whose computed
// output is empty right now is not a move (a bot would spend Vivi's
// once-per-turn activation for nothing), and it is one again as soon
// as the output is not empty.

func TestAPowerZeroViviIsNotAManaMove(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	vivi := battlefieldCard(g, active, game.Card{
		Name: "Vivi Ornitier", TypeLine: "Legendary Creature — Wizard",
		OracleID: oracleViviOrnitier, Power: 0, Toughness: 3,
	})
	advanceTo(t, g, game.StepPrecombatMain)

	if hasManaMoveFrom(legal.EnumerateFor(g, active.ID), vivi) {
		t.Error("offered a power-0 Vivi's mana ability, which adds nothing")
	}

	g.WithWriteLock(func() { _ = g.AddCounterForEffect(vivi, game.CounterPlusOne, 2) })
	if !hasManaMoveFrom(legal.EnumerateFor(g, active.ID), vivi) {
		t.Error("a power-2 Vivi's mana ability should be a move")
	}
}
