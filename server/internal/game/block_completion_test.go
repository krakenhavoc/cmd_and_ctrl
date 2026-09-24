package game

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// block_completion_test.go — #1279. Each defending player's CR 509.1
// block declaration has a completion point, so "has not declared yet"
// (pending) and "declared no blocks" (declared, none) are two states.

// blockersDeclaredEvents returns the EventBlockersDeclared events
// naming `seat` as the declaring defender.
func blockersDeclaredEvents(g *Game, seat uuid.UUID) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == EventBlockersDeclared && ev.Actor == seat {
			out = append(out, ev)
		}
	}
	return out
}

// A defender with no creature at all has no legal block, so their
// declaration — none — is complete the moment the step begins.
func TestDefenderWithNoLegalBlockHasDeclaredNoneAtStepStart(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
	declareAttacks(t, g, attacker)

	def := g.Seats[1].ID
	if got := g.BlockDeclarationStatusOf(def); got != BlockDeclarationDeclared {
		t.Fatalf("status = %q, want declared", got)
	}
	evs := blockersDeclaredEvents(g, def)
	if len(evs) != 1 || evs[0].Amount != 0 {
		t.Fatalf("want one EventBlockersDeclared with Amount 0 (declared, none), got %+v", evs)
	}
	g.WithWriteLock(func() {
		if !g.UnblockedAttackerForEffect(attacker) {
			t.Error("with the declaration complete and nothing blocking, the attacker is unblocked")
		}
	})
	// The active player's own seat is not defending.
	if got := g.BlockDeclarationStatusOf(g.Seats[0].ID); got != BlockDeclarationNone {
		t.Errorf("the attacking seat's status = %q, want none", got)
	}
}

// A defender with a legal block who has not acted is PENDING: the
// attacker is neither blocked nor unblocked, nothing is announced, and
// the wire lists the seat as still declaring.
func TestDefenderWhoHasNotActedIsPending(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
	pushCombatant(t, g, g.Seats[1], "Wall", 0, 4)
	declareAttacks(t, g, attacker)

	def := g.Seats[1].ID
	if got := g.BlockDeclarationStatusOf(def); got != BlockDeclarationPending {
		t.Fatalf("status = %q, want pending", got)
	}
	if n := len(blockersDeclaredEvents(g, def)); n != 0 {
		t.Fatalf("a pending defender has announced nothing: %d events", n)
	}
	g.WithWriteLock(func() {
		if g.UnblockedAttackerForEffect(attacker) {
			t.Error("the attacker read unblocked before its defender declared (the #1279 bug)")
		}
		pending, declared := g.BlockDeclarationSeatsLocked()
		if len(pending) != 1 || pending[0] != 1 || len(declared) != 0 {
			t.Errorf("seats pending=%v declared=%v, want [1] / []", pending, declared)
		}
	})
	if !g.SeatOwesBlockDecision(def) {
		t.Error("a pending defender with a legal block owes the decision (#328)")
	}
}

// finish_blocks with nothing staged is "declared, none": the attacker
// is unblocked, the event carries 0, and the active player — who held
// priority throughout — keeps it.
func TestFinishBlocksWithNothingStagedDeclaresNone(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
	pushCombatant(t, g, g.Seats[1], "Wall", 0, 4)
	declareAttacks(t, g, attacker)
	def := g.Seats[1].ID

	if err := g.FinishBlocks(def); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	if got := g.BlockDeclarationStatusOf(def); got != BlockDeclarationDeclared {
		t.Fatalf("status = %q, want declared", got)
	}
	evs := blockersDeclaredEvents(g, def)
	if len(evs) != 1 || evs[0].Amount != 0 {
		t.Fatalf("want one EventBlockersDeclared with Amount 0, got %+v", evs)
	}
	g.WithWriteLock(func() {
		if !g.UnblockedAttackerForEffect(attacker) {
			t.Error("declared none: the attacker is unblocked")
		}
	})
	if g.SeatOwesBlockDecision(def) {
		t.Error("a declared defender owes nothing, even with an untapped creature at home")
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Errorf("priority holder %d, want the active seat", g.Turn.PriorityHolder)
	}
	// Idempotent.
	if err := g.FinishBlocks(def); err != nil {
		t.Fatalf("second FinishBlocks: %v", err)
	}
	if n := len(blockersDeclaredEvents(g, def)); n != 1 {
		t.Errorf("a second finish announces nothing: %d events", n)
	}
}

// A staged block is announced when its defender's declaration
// COMPLETES — not at an earlier, unrelated priority boundary — and the
// completion event comes after the block events it summarises.
func TestStagedBlockAnnouncesAtCompletionNotBefore(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
	wall := pushCombatant(t, g, g.Seats[1], "Wall", 0, 4)
	declareAttacks(t, g, attacker)
	def := g.Seats[1].ID

	if err := g.DeclareBlocker(wall, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	// An unrelated boundary: a spell resolving in the step would run
	// exactly this.
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	for _, ev := range g.Events {
		if ev.Kind == EventBlock || ev.Kind == EventBecomesBlocked {
			t.Fatalf("a pending defender's staged block was announced at an unrelated boundary: %+v", ev)
		}
	}

	if err := g.FinishBlocks(def); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	var blockSeq, becameSeq, declaredSeq uint64
	for _, ev := range g.Events {
		switch {
		case ev.Kind == EventBlock && ev.CardID == wall:
			blockSeq = ev.Seq
		case ev.Kind == EventBecomesBlocked && ev.CardID == attacker:
			becameSeq = ev.Seq
		case ev.Kind == EventBlockersDeclared && ev.Actor == def:
			declaredSeq = ev.Seq
			if ev.Amount != 1 {
				t.Errorf("Amount = %d, want 1 blocker", ev.Amount)
			}
		}
	}
	if blockSeq == 0 || becameSeq == 0 || declaredSeq == 0 {
		t.Fatalf("completion announces the block, the blocked attacker and the declaration: %d %d %d", blockSeq, becameSeq, declaredSeq)
	}
	if !(blockSeq < declaredSeq && becameSeq < declaredSeq) {
		t.Errorf("the declaration event must follow the blocks it summarises: block %d, became %d, declared %d", blockSeq, becameSeq, declaredSeq)
	}
	g.WithWriteLock(func() {
		if g.UnblockedAttackerForEffect(attacker) {
			t.Error("a blocked attacker is not unblocked")
		}
	})
}

// The defender PASSING is the manual table's "done": it completes the
// declaration, and since that was the last one pending the ACTIVE
// player gets priority back (CR 509.2, 117.3a) instead of the step
// wrapping on the pass.
func TestDefenderPassCompletesTheDeclarationAndReturnsPriorityToTheActivePlayer(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
	pushCombatant(t, g, g.Seats[1], "Wall", 0, 4)
	declareAttacks(t, g, attacker)
	def := g.Seats[1].ID

	if err := g.PassPriority(); err != nil { // active player, before blocks
		t.Fatal(err)
	}
	if g.Turn.PriorityHolder != 1 {
		t.Fatalf("priority should be with the defender, at %d", g.Turn.PriorityHolder)
	}
	if err := g.PassPriority(); err != nil { // the defender: done, no blocks
		t.Fatal(err)
	}
	if g.Turn.Step != StepDeclareBlockers {
		t.Fatalf("the declaring pass must not end the step; now %s", g.Turn.Step)
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Fatalf("after the declaration the active player receives priority; holder %d", g.Turn.PriorityHolder)
	}
	if got := g.BlockDeclarationStatusOf(def); got != BlockDeclarationDeclared {
		t.Fatalf("status = %q, want declared", got)
	}
	// And from there the ordinary round ends the step.
	passUntilStep(t, g, StepCombatDamage)
	if got := StartingLife - g.Seats[1].Life; got != 2 {
		t.Errorf("the unblocked Bear dealt %d, want 2", got)
	}
}

// Leaving the step completes whatever is still pending: AdvanceStep is
// the sandbox's skip-ahead, and the declaration it cuts short is the
// staged one.
func TestAdvanceStepCompletesAPendingDeclaration(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
	wall := pushCombatant(t, g, g.Seats[1], "Wall", 0, 4)
	declareAttacks(t, g, attacker)
	def := g.Seats[1].ID
	if err := g.DeclareBlocker(wall, attacker); err != nil {
		t.Fatal(err)
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	evs := blockersDeclaredEvents(g, def)
	if len(evs) != 1 || evs[0].Amount != 1 {
		t.Fatalf("leaving the step completes the staged declaration: %+v", evs)
	}
	if !g.blockedAttackers[attacker] {
		t.Error("the staged block was locked in as the step ended")
	}
}

// Two defenders: the first to finish does not end the turn-based
// action, so priority does not move; the second does.
func TestSecondOfTwoDefendersClosesTheDeclaration(t *testing.T) {
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
	// Priority is moved off the active player so the close is visible.
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	if g.Turn.PriorityHolder != 1 {
		t.Fatalf("one defender still pending: priority stays put, at %d", g.Turn.PriorityHolder)
	}
	g.WithWriteLock(func() {
		if !g.UnblockedAttackerForEffect(a2) {
			t.Error("a2's defender has declared none, so a2 is unblocked")
		}
		if g.UnblockedAttackerForEffect(a1) {
			t.Error("a1's defender is still declaring")
		}
	})
	if err := g.FinishBlocks(g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Fatalf("the last declaration closes the action: priority to the active seat, at %d", g.Turn.PriorityHolder)
	}
}

// Once declared, a defender is offered nothing more — the enumerator's
// generator answers empty — but a late block by hand is still taken
// (the sandbox allowance) and announced at the next boundary.
func TestDeclaredDefenderIsOfferedNoBlocksButMayStillBlockByHand(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
	wall := pushCombatant(t, g, g.Seats[1], "Wall", 0, 4)
	declareAttacks(t, g, attacker)
	def := g.Seats[1].ID
	if err := g.FinishBlocks(def); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if opts := g.BlockOptionsLocked(def, 4); len(opts) != 0 {
			t.Errorf("a declared defender is offered %d blocks, want none", len(opts))
		}
	})
	if err := g.DeclareBlocker(wall, attacker); err != nil {
		t.Fatalf("the late block by hand was refused: %v", err)
	}
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if !g.blockedAttackers[attacker] {
		t.Error("the late block was announced at the next boundary")
	}
}

func TestFinishBlocksRefusals(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
	pushCombatant(t, g, g.Seats[1], "Wall", 0, 4)
	if err := g.FinishBlocks(g.Seats[1].ID); !errors.Is(err, ErrWrongStep) {
		t.Errorf("outside declare_blockers: %v, want ErrWrongStep", err)
	}
	declareAttacks(t, g, attacker)
	if err := g.FinishBlocks(g.Seats[0].ID); !errors.Is(err, ErrNotDefending) {
		t.Errorf("the attacking seat: %v, want ErrNotDefending", err)
	}
	if err := g.FinishBlocks(uuid.New()); !errors.Is(err, ErrPlayerNotFound) {
		t.Errorf("an unknown seat: %v, want ErrPlayerNotFound", err)
	}
}

// Undo: the completion rewinds with the declaration, in both the
// in-memory undo (Clone / RestoreFrom) and the persisted snapshot.
func TestBlockDeclarationCompletionRoundTripsUndoAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
	pushCombatant(t, g, g.Seats[1], "Wall", 0, 4)
	declareAttacks(t, g, attacker)
	def := g.Seats[1].ID

	before := g.Clone()
	if err := g.FinishBlocks(def); err != nil {
		t.Fatal(err)
	}
	after := g.Clone()

	// A persisted snapshot of the declared state restores declared.
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatal(err)
	}
	if got := restored.BlockDeclarationStatusOf(def); got != BlockDeclarationDeclared {
		t.Errorf("snapshot round trip: status %q, want declared", got)
	}

	// Undo back past the finish: pending again, and the attacker is
	// not unblocked.
	g.RestoreFrom(before)
	if got := g.BlockDeclarationStatusOf(def); got != BlockDeclarationPending {
		t.Fatalf("after undo: status %q, want pending", got)
	}
	g.WithWriteLock(func() {
		if g.UnblockedAttackerForEffect(attacker) {
			t.Error("after undo the attacker must not read unblocked")
		}
	})
	// Redo.
	g.RestoreFrom(after)
	if got := g.BlockDeclarationStatusOf(def); got != BlockDeclarationDeclared {
		t.Errorf("after redo: status %q, want declared", got)
	}
}

// The next combat starts from nothing: clearCombatLocked forgets which
// defenders declared.
func TestCombatEndForgetsWhoDeclared(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
	pushCombatant(t, g, g.Seats[1], "Wall", 0, 4)
	declareAttacks(t, g, attacker)
	if err := g.FinishBlocks(g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	if err := g.ClearCombat(); err != nil {
		t.Fatal(err)
	}
	if g.blocksDeclared != nil {
		t.Errorf("clear combat left %v", g.blocksDeclared)
	}
}
