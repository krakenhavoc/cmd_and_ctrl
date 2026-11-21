package testsbr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// Leyline Tyrant dies: the controller picks any amount of {R} up to the red
// they can pay, pays it, and a reflexive trigger deals that much damage to
// the target picked as it goes on the stack.
func TestLeylineTyrantDiesPayRedForThatMuchDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	tyrant := effectstest.PushCatalogPermanent(g, me.ID, "Leyline Tyrant", "Creature — Dragon", effectstest.MkTyrantOracle, false)
	effectstest.FillPoolColored(me, "R", 3)
	effectstest.FillPool(me, 2) // generic mana is not red: it can't be paid
	before := opp.Life

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tyrant) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	pa := effectstest.AnswerNumber(t, g, me.ID, 2)
	if pa.Min != 0 || pa.Max != 3 {
		t.Fatalf("X prompt = %+v, want 0..3 (only red counts)", pa)
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if got := effectstest.RedInPool(me); got != 1 {
		t.Errorf("red left in the pool = %d, want 1 (paid {R}{R})", got)
	}
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	if opp.Life != before {
		t.Fatal("damage landed before the reflexive trigger resolved")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 2 {
		t.Errorf("damage = %d, want 2", got)
	}
}

// Paying nothing creates no trigger and deals no damage.
func TestLeylineTyrantDiesPayingNothingDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	tyrant := effectstest.PushCatalogPermanent(g, me.ID, "Leyline Tyrant", "Creature — Dragon", effectstest.MkTyrantOracle, false)
	effectstest.FillPoolColored(me, "R", 3)
	before := opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tyrant) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerNumber(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Errorf("life %d → %d, want no damage", before, opp.Life)
	}
	if got := effectstest.RedInPool(me); got != 3 {
		t.Errorf("red in the pool = %d, want 3 untouched", got)
	}
}

// With no red to pay the controller isn't asked at all.
func TestLeylineTyrantDiesWithoutRedAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	tyrant := effectstest.PushCatalogPermanent(g, me.ID, "Leyline Tyrant", "Creature — Dragon", effectstest.MkTyrantOracle, false)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tyrant) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PayAmountPrompt(g, me.ID) != nil {
		t.Error("asked for X with no red mana to pay")
	}
}
