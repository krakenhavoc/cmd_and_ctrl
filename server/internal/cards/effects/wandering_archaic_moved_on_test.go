package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// wandering_archaic_moved_on_test.go — #4276. Wandering Archaic's {2}
// question is a pay_unless, and a pay_unless does not stop the table
// (choice_gate.go, ADR 0018 §6), so the caster can leave it open while
// the spell resolves and play moves on. The copy that follows a decline
// must not depend on the resolving slot (Game.resolving), which lives
// for one event batch. CR 608.2h: "copy that spell" names the spell and
// does not target it, so it uses last-known information.

// archaicTableWithOpenPayUnless casts a Bolt at the Archaic's controller
// on the opponent's turn and returns the table with the {2} question
// open and the Bolt already resolved (3 damage taken).
func archaicTableWithOpenPayUnless(t *testing.T) (g *game.Game, lifeBefore int) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Wandering Archaic", "Creature — Avatar", wanderingArchaicOracle, false)
	for g.Turn.ActiveSeat != 1 || (g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain) {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	lifeBefore = me.Life
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatalf("no pay-unless prompt for the caster")
	}
	if got := lifeBefore - me.Life; got != 3 {
		t.Fatalf("the bolt has dealt %d, want 3: the question is not open over a resolved spell", got)
	}
	return g, lifeBefore
}

// answerDeclineThenCopy declines the {2} and says yes to the copy.
func answerDeclineThenCopy(t *testing.T, g *game.Game) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	answerPayUnless(t, g, opp.ID, false)
	answerMayChoice(t, g, me.ID, true)
	settleThroughCopyRetarget(t, g, me.ID, me.ID)
}

// The caster leaves the question open while the table walks into the next
// step (beginEventBatchLocked clears the resolving slot there), then
// declines. The copy is still made, from the Bolt as it last was on the
// stack, and it is not a targeted effect, so nothing fizzles it.
func TestWanderingArchaicCopyAfterTheTableMovedOnAStep(t *testing.T) {
	g, before := archaicTableWithOpenPayUnless(t)
	me, opp := g.Seats[0], g.Seats[1]
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep with an open pay-unless: %v", err)
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatalf("the pay-unless prompt went away when the step advanced")
	}
	answerDeclineThenCopy(t, g)
	if got := before - me.Life; got != 6 {
		t.Errorf("took %d damage, want 6 (bolt + copy)", got)
	}
}

// The same, across a whole turn boundary: the per-turn last-known
// records are cleared there, except what an open prompt still names.
func TestWanderingArchaicCopyAfterTheTurnMovedOn(t *testing.T) {
	g, before := archaicTableWithOpenPayUnless(t)
	me, opp := g.Seats[0], g.Seats[1]
	turn := g.Turn.Seq
	for i := 0; g.Turn.Seq == turn; i++ {
		if i > 40 {
			t.Fatalf("the turn never ended")
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatalf("the pay-unless prompt went away over the turn boundary")
	}
	answerDeclineThenCopy(t, g)
	if got := before - me.Life; got != 6 {
		t.Errorf("took %d damage, want 6 (bolt + copy)", got)
	}
}

// A save taken while the question is open over a resolved spell, after
// the table has moved on, restores to a table that still copies: the held
// record is carried by the snapshot (GameSnapshot.LastKnownStack), no new
// field.
func TestWanderingArchaicCopyAfterAMoveOnAndARestore(t *testing.T) {
	g, before := archaicTableWithOpenPayUnless(t)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	g = restoreRoundTrip(t, g, true)
	answerDeclineThenCopy(t, g)
	if got := before - g.Seats[0].Life; got != 6 {
		t.Errorf("took %d damage, want 6 (bolt + copy)", got)
	}
}
