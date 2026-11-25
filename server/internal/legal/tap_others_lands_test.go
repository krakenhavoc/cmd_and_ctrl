package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// tap_others_lands_test.go — the enumerator half of #6204 (Piracy): an
// opponent's land is a mana move for the player Piracy lets tap it, and
// for nobody else, and the move is one the dispatcher accepts (#544).
func TestPiracyOffersAnOpponentsLandAsAManaMove(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	controller := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bystander := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	mountain := battlefieldCard(g, controller, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	advanceTo(t, g, game.StepPrecombatMain)

	if moves := movesFrom(legal.EnumerateFor(g, active.ID), mountain, legal.KindMana); len(moves) != 0 {
		t.Fatalf("without Piracy: offered %v", labels(moves))
	}
	g.WithWriteLock(func() { g.TapLandsYouDontControlThisTurnForEffect(uuid.Nil, active.ID, "Piracy") })
	got := movesFrom(legal.EnumerateFor(g, active.ID), mountain, legal.KindMana)
	if len(got) != 1 {
		t.Fatalf("under Piracy: want one mana move on the opponent's Mountain, got %v", labels(got))
	}
	if got[0].Player != active.ID {
		t.Errorf("move player = %v, want the activator %v", got[0].Player, active.ID)
	}
	if moves := movesFrom(legal.EnumerateFor(g, bystander.ID), mountain, legal.KindMana); len(moves) != 0 {
		t.Errorf("a bystander without Piracy was offered %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, got)
}
