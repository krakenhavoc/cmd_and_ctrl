package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Leyline Tyrant dies: the controller picks any amount of {R} up to the red
// they can pay, pays it, and a reflexive trigger deals that much damage to
// the target picked as it goes on the stack.
func TestLeylineTyrantDiesPayRedForThatMuchDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	tyrant := pushCatalogPermanent(g, me.ID, "Leyline Tyrant", "Creature — Dragon", mkTyrantOracle, false)
	fillPoolColored(me, "R", 3)
	fillPool(me, 2) // generic mana is not red: it can't be paid
	before := opp.Life

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tyrant) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	pa := answerNumber(t, g, me.ID, 2)
	if pa.Min != 0 || pa.Max != 3 {
		t.Fatalf("X prompt = %+v, want 0..3 (only red counts)", pa)
	}
	answerPayUnless(t, g, me.ID, true)
	if got := redInPool(me); got != 1 {
		t.Errorf("red left in the pool = %d, want 1 (paid {R}{R})", got)
	}
	answerPickTargetPlayer(t, g, opp.ID)
	if opp.Life != before {
		t.Fatal("damage landed before the reflexive trigger resolved")
	}
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 2 {
		t.Errorf("damage = %d, want 2", got)
	}
}

// Paying nothing creates no trigger and deals no damage.
func TestLeylineTyrantDiesPayingNothingDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	tyrant := pushCatalogPermanent(g, me.ID, "Leyline Tyrant", "Creature — Dragon", mkTyrantOracle, false)
	fillPoolColored(me, "R", 3)
	before := opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tyrant) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	answerNumber(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Errorf("life %d → %d, want no damage", before, opp.Life)
	}
	if got := redInPool(me); got != 3 {
		t.Errorf("red in the pool = %d, want 3 untouched", got)
	}
}

// With no red to pay the controller isn't asked at all.
func TestLeylineTyrantDiesWithoutRedAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	tyrant := pushCatalogPermanent(g, me.ID, "Leyline Tyrant", "Creature — Dragon", mkTyrantOracle, false)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tyrant) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if payAmountPrompt(g, me.ID) != nil {
		t.Error("asked for X with no red mana to pay")
	}
}

// redInPool counts the red mana in a player's pool.
func redInPool(p *game.Player) int {
	n := 0
	for _, c := range poolColors(p) {
		if c == "R" {
			n++
		}
	}
	return n
}
