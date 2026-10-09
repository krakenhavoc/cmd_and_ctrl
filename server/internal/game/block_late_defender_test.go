package game

import (
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// block_late_defender_test.go — #2021, CR 509.1 / 509.1h / 508.5.
// Declaring blockers is ONE turn-based action, taken as the declare
// blockers step begins. A player who becomes a defending player after
// it is over — an attack reselected onto them (CR 508.7a), a creature
// put onto the battlefield attacking them (CR 508.4) — was not
// defending when it happened, so they never declare: an attacker
// pointed at them stays as the declaration left it.

// lateDefenderTable is a four-seat table in declare_blockers: seat 0's
// 3/3 attacked seat 1, seat 1 has finished declaring (none), and the
// attack has since been reselected onto seat 2, who has an untapped
// 0/4 and was not defending when blockers were declared.
func lateDefenderTable(t *testing.T) (g *Game, attacker, lateWall uuid.UUID) {
	t.Helper()
	g = newActiveGameWithSeats(t, 4)
	attacker = pushCombatant(t, g, g.Seats[0], "Redirected", 3, 3)
	pushCombatant(t, g, g.Seats[1], "First Defender's Wall", 0, 4)
	lateWall = pushCombatant(t, g, g.Seats[2], "Late Defender's Wall", 0, 4)
	declareAttacks(t, g, attacker)
	if err := g.FinishBlocks(g.Seats[1].ID); err != nil {
		t.Fatalf("seat 1 declares no blocks: %v", err)
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Fatalf("setup: the declaration is over and the active player has priority; holder %d", g.Turn.PriorityHolder)
	}
	if err := reselect(g, attacker, g.Seats[2].ID); err != nil {
		t.Fatalf("ReselectAttackTargetForEffect: %v", err)
	}
	return g, attacker, lateWall
}

// The issue's shape: the new defending player is not asked to declare,
// is offered nothing, is refused a block, and the attacker — unblocked
// by the declaration that happened — hits them.
func TestPlayerWhoBecomesDefendingAfterTheDeclarationDoesNotDeclare(t *testing.T) {
	g, attacker, lateWall := lateDefenderTable(t)
	late := g.Seats[2].ID

	if got := g.BlockDeclarationStatusOf(late); got != BlockDeclarationDeclared {
		t.Errorf("the late defender's status = %q, want declared", got)
	}
	var pending, declared []int
	g.WithWriteLock(func() { pending, declared = g.BlockDeclarationSeatsLocked() })
	if len(pending) != 0 || !slices.Contains(declared, 2) {
		t.Errorf("wire: block_pending_seats %v, blocks_declared_seats %v; want none pending and seat 2 declared", pending, declared)
	}
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if opts := g.BlockOptionsLocked(late, 4); len(opts) != 0 {
			t.Errorf("the generator offers the late defender %d blocks, want none", len(opts))
		}
		if !g.UnblockedAttackerForEffect(attacker) {
			t.Error("the declaration left the attacker unblocked; it must still read unblocked")
		}
	})
	if g.SeatOwesBlockDecision(late) {
		t.Error("the #328 signal says the late defender owes a block decision")
	}
	var br *BlockRefusedError
	if err := g.DeclareBlocker(lateWall, attacker); !errors.As(err, &br) || br.Reason != BlockReasonBlocksDeclared {
		t.Fatalf("the late defender blocks: %v, want blocks_declared", err)
	}
	if c := findCard(g, lateWall); c.BlockingTarget != uuid.Nil {
		t.Errorf("the refused block was stored: blocking %v", c.BlockingTarget)
	}
	// finish_blocks is the idempotent no-op it is for any declared seat:
	// no announcement, and priority stays where it was.
	holder := g.Turn.PriorityHolder
	if err := g.FinishBlocks(late); err != nil {
		t.Errorf("FinishBlocks for the late defender: %v, want the declared no-op", err)
	}
	if g.Turn.PriorityHolder != holder {
		t.Errorf("FinishBlocks moved priority from %d to %d", holder, g.Turn.PriorityHolder)
	}

	// Their pass is an ordinary pass, not a declaration that hands the
	// active player priority back.
	for _, want := range []int{1, 2, 3} {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		if g.Turn.Step != StepDeclareBlockers || g.Turn.PriorityHolder != want {
			t.Fatalf("after a pass: %s, holder %d; want declare_blockers, holder %d", g.Turn.Step, g.Turn.PriorityHolder, want)
		}
	}
	before := g.Seats[2].Life
	passUntilStep(t, g, StepEndCombat)
	if evs := blockersDeclaredEvents(g, late); len(evs) != 0 {
		t.Errorf("the late defender was announced as declaring: %+v", evs)
	}
	if got := before - g.Seats[2].Life; got != 3 {
		t.Errorf("the late defender lost %d life, want 3 from the unblocked attacker", got)
	}
}

// The other way in: a creature put onto the battlefield attacking a
// player nothing was attacking when blockers were declared (CR 508.4).
func TestCreatureEnteringAttackingANewPlayerAfterTheDeclarationCannotBeBlocked(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	me, first, late := g.Seats[0], g.Seats[1], g.Seats[2]
	attacker := pushCombatant(t, g, me, "Grizzly Bears", 2, 2)
	ninja := pushHandCreature(t, g, me, "Ninja", 3, 3)
	wall := pushCombatant(t, g, late, "Wall", 0, 4)
	declareAttacks(t, g, attacker)
	if err := g.FinishBlocks(first.ID); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(ninja, HandEntryOptions{
			Controller: me.ID,
			Tapped:     true,
			Attacking:  late.ID,
		}); err != nil {
			t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
		}
	})
	if got := g.BlockDeclarationStatusOf(late.ID); got != BlockDeclarationDeclared {
		t.Errorf("status of the player the ninja attacks = %q, want declared", got)
	}
	var br *BlockRefusedError
	if err := g.DeclareBlocker(wall, ninja); !errors.As(err, &br) || br.Reason != BlockReasonBlocksDeclared {
		t.Fatalf("blocking the ninja: %v, want blocks_declared", err)
	}
	before := late.Life
	passUntilStep(t, g, StepEndCombat)
	if got := before - late.Life; got != 3 {
		t.Errorf("the ninja's player lost %d life, want 3", got)
	}
}

// The record is "every player defending at that moment has declared":
// not set while one is still declaring, set by the last one.
func TestTheDeclarationClosesWithTheLastDefender(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	a1 := pushCombatant(t, g, g.Seats[0], "Bear A", 2, 2)
	a2 := pushCombatant(t, g, g.Seats[0], "Bear B", 2, 2)
	pushCombatant(t, g, g.Seats[1], "Wall 1", 0, 4)
	pushCombatant(t, g, g.Seats[2], "Wall 2", 0, 4)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(a1, g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(a2, g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	if err := g.FinishBlocks(g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	if g.blockDeclarationClosed {
		t.Fatal("closed while seat 2 is still declaring")
	}
	if got := g.BlockDeclarationStatusOf(g.Seats[2].ID); got != BlockDeclarationPending {
		t.Fatalf("seat 2 = %q, want pending", got)
	}
	if err := g.FinishBlocks(g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	if !g.blockDeclarationClosed {
		t.Fatal("the last declaration did not close the declaration")
	}
}

// Undo, the persisted snapshot and the next combat: the record rewinds
// with the declaration and is forgotten with combat.
func TestClosedDeclarationRoundTripsUndoAndSnapshot(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Redirected", 3, 3)
	pushCombatant(t, g, g.Seats[1], "First Defender's Wall", 0, 4)
	pushCombatant(t, g, g.Seats[2], "Late Defender's Wall", 0, 4)
	declareAttacks(t, g, attacker)
	before := g.Clone()
	if err := g.FinishBlocks(g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	if err := reselect(g, attacker, g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	after := g.Clone()
	late := g.Seats[2].ID

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	if !snap.BlockDeclarationClosed {
		t.Error("the snapshot does not record the closed declaration")
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatal(err)
	}
	if got := restored.BlockDeclarationStatusOf(late); got != BlockDeclarationDeclared {
		t.Errorf("snapshot round trip: the late defender reads %q, want declared", got)
	}

	g.RestoreFrom(before)
	if g.blockDeclarationClosed {
		t.Error("undo back past the declaration kept it closed")
	}
	g.RestoreFrom(after)
	if got := g.BlockDeclarationStatusOf(late); got != BlockDeclarationDeclared {
		t.Errorf("after redo: the late defender reads %q, want declared", got)
	}

	if err := g.ClearCombat(); err != nil {
		t.Fatal(err)
	}
	if g.blockDeclarationClosed {
		t.Error("clear combat kept the closed declaration")
	}
}
