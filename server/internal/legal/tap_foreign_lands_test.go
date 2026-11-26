package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// tap_foreign_lands_test.go — the enumerator half of #6204 (ADR 0106 §1
// amendment 2026-11-25): under "Until end of turn, you may tap lands you
// don't control for mana" (Piracy) another player's land is a mana move
// for the permission's holder, the move is one the dispatcher accepts
// (#544), and nobody else is offered it. CR 602.2.
func TestForeignLandIsAManaMoveUnderThePermission(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	controller := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bystander := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	mountain := battlefieldCard(g, controller, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	advanceTo(t, g, game.StepPrecombatMain)

	if moves := movesFrom(legal.EnumerateFor(g, active.ID), mountain, legal.KindMana); len(moves) != 0 {
		t.Fatalf("without the permission: offered %v", labels(moves))
	}
	g.WithWriteLock(func() { g.AllowTapForeignLandsThisTurnForEffect(active.ID) })
	got := movesFrom(legal.EnumerateFor(g, active.ID), mountain, legal.KindMana)
	if len(got) != 1 {
		t.Fatalf("with the permission: want one mana move on the Mountain, got %v", labels(got))
	}
	if moves := movesFrom(legal.EnumerateFor(g, bystander.ID), mountain, legal.KindMana); len(moves) != 0 {
		t.Errorf("a bystander without the permission: offered %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, got)
}
