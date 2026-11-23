package testsu

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// wandering_archaic_moved_on_test.go — #4276, revisited by #5793.
// Wandering Archaic's {2} question is a pay_unless. #4276 was written when
// a pay_unless did not stop the table (ADR 0018 §6), so the caster could
// leave it open while the spell resolved and play moved on, and the copy
// that follows a decline had to work from last-known information.
//
// Since #5793 the question is owed in the step it is asked in (CR 118.12:
// the payment is part of the trigger's resolution), so the table waits for
// it with the spell still on the stack. What #4276 still guards is that
// the copy does not depend on the resolving slot (Game.resolving), which
// lives for one event batch, while the answer comes in a later one, or
// after a restore.

// archaicTableWithOpenPayUnless casts a Bolt at the Archaic's controller
// on the opponent's turn and returns the table with the {2} question open
// over the Bolt, which has not resolved yet.
func archaicTableWithOpenPayUnless(t *testing.T) (g *game.Game, lifeBefore int) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Wandering Archaic", "Creature — Avatar", effectstest.WanderingArchaicOracle, false)
	for g.Turn.ActiveSeat != 1 || (g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain) {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	lifeBefore = me.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatalf("no pay-unless prompt for the caster")
	}
	if got := lifeBefore - me.Life; got != 0 {
		t.Fatalf("the bolt has dealt %d with the question open, want 0: the table did not wait", got)
	}
	return g, lifeBefore
}

// answerDeclineThenCopy declines the {2}, says yes to the copy, and
// settles the stack.
func answerDeclineThenCopy(t *testing.T, g *game.Game) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.SettleThroughCopyRetarget(t, g, me.ID, me.ID)
}

// The table cannot walk on with the question open (#5793), and the copy is
// made once it is answered, in a later event batch than the one that
// asked.
func TestWanderingArchaicHoldsTheTableWhileTheCasterDecides(t *testing.T) {
	g, before := archaicTableWithOpenPayUnless(t)
	me, opp := g.Seats[0], g.Seats[1]
	step := g.Turn.Step
	if _, err := g.AdvanceStep(); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("AdvanceStep with the {2} question open = %v, want ErrChoicePending", err)
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("PassPriority with the {2} question open = %v, want ErrChoicePending", err)
	}
	if g.Turn.Step != step || !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatalf("the table moved (step %v) or lost the question", g.Turn.Step)
	}
	answerDeclineThenCopy(t, g)
	if got := before - me.Life; got != 6 {
		t.Errorf("took %d damage, want 6 (bolt + copy)", got)
	}
}

// A save taken while the question is open restores to a table that still
// holds and still copies.
func TestWanderingArchaicCopyAfterARestore(t *testing.T) {
	g, before := archaicTableWithOpenPayUnless(t)
	g = effectstest.RestoreRoundTrip(t, g, true)
	if _, err := g.AdvanceStep(); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("restored AdvanceStep with the {2} question open = %v, want ErrChoicePending", err)
	}
	answerDeclineThenCopy(t, g)
	if got := before - g.Seats[0].Life; got != 6 {
		t.Errorf("took %d damage, want 6 (bolt + copy)", got)
	}
}
