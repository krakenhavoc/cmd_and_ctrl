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
// is unblocked, the event carries 0, and the active player receives the
// priority that was parked while the defender declared (#1501).
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

// #1501 / CR 509.1: the declaration is taken BEFORE anyone receives
// priority. With a defender still declaring the step parks priority —
// nobody may pass, and the active player has no pre-declaration window
// — until the defender finishes; then the ACTIVE player receives
// priority (CR 509.2, 117.3a), which is the post-block window, and the
// ordinary round ends the step.
func TestPriorityIsParkedUntilTheDefenderDeclares(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
	pushCombatant(t, g, g.Seats[1], "Wall", 0, 4)
	declareAttacks(t, g, attacker)
	def := g.Seats[1].ID

	if g.Turn.PriorityHolder != NoPriority {
		t.Fatalf("priority must be parked while the defender declares; holder %d", g.Turn.PriorityHolder)
	}
	if err := g.PassPriority(); !errors.Is(err, ErrNoPriority) {
		t.Fatalf("a pass before the declaration: %v, want ErrNoPriority", err)
	}
	if err := g.FinishBlocks(def); err != nil { // the defender: done, no blocks
		t.Fatal(err)
	}
	if g.Turn.Step != StepDeclareBlockers {
		t.Fatalf("finishing the declaration must not end the step; now %s", g.Turn.Step)
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Fatalf("after the declaration the active player receives priority; holder %d", g.Turn.PriorityHolder)
	}
	if got := g.BlockDeclarationStatusOf(def); got != BlockDeclarationDeclared {
		t.Fatalf("status = %q, want declared", got)
	}
	// The post-block window is real: the active player's pass hands
	// priority to the defender rather than ending the step.
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	if g.Turn.Step != StepDeclareBlockers || g.Turn.PriorityHolder != 1 {
		t.Fatalf("the active player's post-block pass goes to the defender: %s, holder %d", g.Turn.Step, g.Turn.PriorityHolder)
	}
	// And from there the ordinary round ends the step.
	passUntilStep(t, g, StepCombatDamage)
	if got := StartingLife - g.Seats[1].Life; got != 2 {
		t.Errorf("the unblocked Bear dealt %d, want 2", got)
	}
}

// Completion point 3 is kept for the one shape in which a defender
// still declaring holds priority: a restore point written before #1501
// mid-step. (A player who became a defending player after the
// declaration closed is not declaring at all, #2021.) Their PASS is still "done", and the last one hands the active
// player priority instead of wrapping the step.
func TestPendingDefenderHoldingPriorityCompletesByPassing(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
	pushCombatant(t, g, g.Seats[1], "Wall", 0, 4)
	declareAttacks(t, g, attacker)
	def := g.Seats[1].ID
	g.WithWriteLock(func() { g.Turn.PriorityHolder = 1 })

	if err := g.PassPriority(); err != nil { // the defender: done, no blocks
		t.Fatal(err)
	}
	if g.Turn.Step != StepDeclareBlockers || g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Fatalf("the declaring pass returns priority to the active player in the step: %s, holder %d", g.Turn.Step, g.Turn.PriorityHolder)
	}
	if got := g.BlockDeclarationStatusOf(def); got != BlockDeclarationDeclared {
		t.Fatalf("status = %q, want declared", got)
	}
}

// A defender with nothing to block with declares "none" as the step
// begins (#1279), so there is nothing to wait for and the active player
// receives priority as the step begins — no parking, no stall.
func TestNoParkingWhenNoDefenderHasABlock(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
	tapped := pushCombatant(t, g, g.Seats[1], "Tapped Wall", 0, 4)
	g.WithWriteLock(func() { findCard(g, tapped).Tapped = true })
	declareAttacks(t, g, attacker)

	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Fatalf("nobody is declaring, so the active player holds priority; holder %d", g.Turn.PriorityHolder)
	}
	if got := g.BlockDeclarationStatusOf(g.Seats[1].ID); got != BlockDeclarationDeclared {
		t.Fatalf("a defender whose only creature is tapped has declared none: %q", got)
	}
	passUntilStep(t, g, StepCombatDamage)
	if got := StartingLife - g.Seats[1].Life; got != 2 {
		t.Errorf("the unblocked Bear dealt %d, want 2", got)
	}
}

// One defender with nothing to block with and one still declaring: the
// table waits on the second alone, and their finish ends the action.
func TestParkedOnlyForTheDefenderStillDeclaring(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	a1 := pushCombatant(t, g, g.Seats[0], "Bear A", 2, 2)
	a2 := pushCombatant(t, g, g.Seats[0], "Bear B", 2, 2)
	pushCombatant(t, g, g.Seats[1], "Wall 1", 0, 4)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(a1, g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(a2, g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)

	if got := g.BlockDeclarationStatusOf(g.Seats[2].ID); got != BlockDeclarationDeclared {
		t.Fatalf("seat 2 has no creature and declared none as the step began: %q", got)
	}
	if g.Turn.PriorityHolder != NoPriority {
		t.Fatalf("seat 1 is still declaring, so priority is parked; holder %d", g.Turn.PriorityHolder)
	}
	if err := g.FinishBlocks(g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Fatalf("the last declaration closes the action: holder %d", g.Turn.PriorityHolder)
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
	// #1501: nobody holds priority while either is declaring.
	if g.Turn.PriorityHolder != NoPriority {
		t.Fatalf("two defenders declaring: priority parked, at %d", g.Turn.PriorityHolder)
	}
	if err := g.FinishBlocks(g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	if g.Turn.PriorityHolder != NoPriority {
		t.Fatalf("one defender still pending: priority stays parked, at %d", g.Turn.PriorityHolder)
	}
	if err := g.PassPriority(); !errors.Is(err, ErrNoPriority) {
		t.Fatalf("a pass while one defender is still declaring: %v, want ErrNoPriority", err)
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

// Three defenders: each finish but the last leaves priority parked, and
// the last hands it to the active player. In Commander every defending
// player declares for themselves (CR 509.1 / 802.4a), in any order.
func TestThirdOfThreeDefendersClosesTheDeclaration(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	var attackers []uuid.UUID
	for i := 1; i <= 3; i++ {
		atk := pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
		pushCombatant(t, g, g.Seats[i], "Wall", 0, 4)
		attackers = append(attackers, atk)
	}
	advanceIntoStep(t, g, StepDeclareAttackers)
	for i, atk := range attackers {
		if err := g.DeclareAttacker(atk, g.Seats[i+1].ID); err != nil {
			t.Fatal(err)
		}
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	for _, seat := range []int{2, 3} {
		if g.Turn.PriorityHolder != NoPriority {
			t.Fatalf("before seat %d finishes: holder %d, want parked", seat, g.Turn.PriorityHolder)
		}
		if err := g.FinishBlocks(g.Seats[seat].ID); err != nil {
			t.Fatal(err)
		}
	}
	if g.Turn.PriorityHolder != NoPriority {
		t.Fatalf("seat 1 is still declaring: holder %d, want parked", g.Turn.PriorityHolder)
	}
	g.WithWriteLock(func() {
		pending, declared := g.BlockDeclarationSeatsLocked()
		if len(pending) != 1 || pending[0] != 1 || len(declared) != 2 {
			t.Errorf("seats pending=%v declared=%v, want [1] / two declared", pending, declared)
		}
	})
	if err := g.FinishBlocks(g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Fatalf("the last declaration closes the action: holder %d", g.Turn.PriorityHolder)
	}
}

// #1501: once declared, a defender is offered nothing more — the
// enumerator's generator answers empty — and the verb agrees: a late
// block is refused with blocks_declared and nothing is stored or
// announced. Before #1501 the verb took it as a sandbox allowance.
func TestDeclaredDefenderIsOfferedNoBlocksAndALateBlockIsRefused(t *testing.T) {
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
	err := g.DeclareBlocker(wall, attacker)
	var br *BlockRefusedError
	if !errors.As(err, &br) || br.Reason != BlockReasonBlocksDeclared {
		t.Fatalf("a late block: %v, want a blocks_declared refusal", err)
	}
	if !errors.Is(err, ErrIllegalBlock) {
		t.Errorf("the refusal does not wrap ErrIllegalBlock: %v", err)
	}
	if br.Blocker != wall || br.Defender != def {
		t.Errorf("refusal names blocker %s / defender %s, want the Wall and the defender", br.Blocker, br.Defender)
	}
	if got := br.Sentence(def); got != "You have already finished declaring blockers this combat." {
		t.Errorf("the defender reads %q", got)
	}
	if got := br.Sentence(g.Seats[0].ID); got != "P2 has already finished declaring blockers this combat." {
		t.Errorf("the table reads %q", got)
	}
	if c := findCard(g, wall); c.BlockingTarget != uuid.Nil {
		t.Errorf("the refused block was stored: blocking %v", c.BlockingTarget)
	}
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if g.blockedAttackers[attacker] {
		t.Error("a refused late block was announced")
	}
}

// #1501: a block that is illegal anyway is refused for what is wrong
// with it, not for being late. A defender whose only creature can't
// block the flier completed "declared, none" as the step began, and the
// flying refusal tells them more than "you have finished declaring".
func TestALateIllegalBlockReportsItsOwnReason(t *testing.T) {
	g := newActiveGame(t)
	flier := pushCombatant(t, g, g.Seats[0], "Bird", 1, 1, "flying")
	bear := pushCombatant(t, g, g.Seats[1], "Bear", 2, 2)
	declareAttacks(t, g, flier)
	if got := g.BlockDeclarationStatusOf(g.Seats[1].ID); got != BlockDeclarationDeclared {
		t.Fatalf("setup: no legal block, so declared at step start; status %q", got)
	}
	var br *BlockRefusedError
	if err := g.DeclareBlocker(bear, flier); !errors.As(err, &br) || br.Reason != BlockReasonFlying {
		t.Fatalf("the flier's refusal: %v, want flying", err)
	}
}

// The ninja's shape, CR 509.1 / 509.1h: a creature put onto the
// battlefield attacking AFTER its defending player has declared cannot
// be blocked — the declaration is over. #1501 makes that an engine
// fact: the generator offers nothing, the #328 signal owes nothing, and
// the verb refuses the block, so the creature connects.
func TestCreatureThatEnteredAttackingAfterTheDeclarationCannotBeBlocked(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushCombatant(t, g, me, "Grizzly Bears", 2, 2)
	ninja := pushHandCreature(t, g, me, "Ninja", 3, 3)
	wall := pushCombatant(t, g, opp, "Wall", 0, 4)
	declareAttacks(t, g, attacker)
	if err := g.FinishBlocks(opp.ID); err != nil { // declared, none
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(ninja, HandEntryOptions{
			Controller: me.ID,
			Tapped:     true,
			Attacking:  opp.ID,
		}); err != nil {
			t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
		}
	})
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if opts := g.BlockOptionsLocked(opp.ID, 4); len(opts) != 0 {
			t.Errorf("the generator offers %d blocks on the arrived ninja, want none", len(opts))
		}
	})
	if g.SeatOwesBlockDecision(opp.ID) {
		t.Error("the #328 signal says the declared defender owes a block decision")
	}
	var br *BlockRefusedError
	if err := g.DeclareBlocker(wall, ninja); !errors.As(err, &br) || br.Reason != BlockReasonBlocksDeclared {
		t.Fatalf("blocking the ninja after the declaration: %v, want blocks_declared", err)
	}
	before := opp.Life
	passUntilStep(t, g, StepEndCombat)
	if got := before - opp.Life; got != 5 {
		t.Errorf("defender lost %d life, want 5 — the Bears and the ninja both connect", got)
	}
}

// A defender who leaves the game while priority is parked for their
// declaration was the one the table was waiting on: the declaration is
// over, and the active player receives priority (#1501).
func TestConcedingPendingDefenderUnparksPriority(t *testing.T) {
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
	if g.Turn.PriorityHolder != NoPriority {
		t.Fatalf("seat 2 is still declaring: holder %d", g.Turn.PriorityHolder)
	}
	if err := g.Concede(g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	if g.State != StateActive || g.Turn.Step != StepDeclareBlockers {
		t.Fatalf("the game goes on in the step: state %s, step %s", g.State, g.Turn.Step)
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Fatalf("with nobody left declaring the active player receives priority; holder %d", g.Turn.PriorityHolder)
	}
}

// The dispatcher's backstop: a sandbox verb that takes every attacker
// out of combat while priority is parked leaves nobody declaring, and
// SettleBlockDeclaration hands the active player priority rather than
// leaving a step nobody can move (#1501).
func TestSettleBlockDeclarationUnparksAStepWithNobodyDeclaring(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
	pushCombatant(t, g, g.Seats[1], "Wall", 0, 4)
	declareAttacks(t, g, attacker)
	if g.Turn.PriorityHolder != NoPriority {
		t.Fatalf("setup: parked; holder %d", g.Turn.PriorityHolder)
	}
	// A no-op while someone is still declaring.
	g.SettleBlockDeclaration()
	if g.Turn.PriorityHolder != NoPriority {
		t.Fatalf("settle unparked a step with a defender still declaring; holder %d", g.Turn.PriorityHolder)
	}
	if err := g.ClearCombat(); err != nil {
		t.Fatal(err)
	}
	g.SettleBlockDeclaration()
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Fatalf("nobody is declaring any more: holder %d, want the active seat", g.Turn.PriorityHolder)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("the table can move again: %v", err)
	}
}

// Undo across the parked step: a clone taken before the step began
// comes back with the active player holding priority in declare
// attackers, and one taken while parked comes back parked with the
// defender still declaring.
func TestParkedPriorityRoundTripsUndo(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
	pushCombatant(t, g, g.Seats[1], "Wall", 0, 4)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	beforeStep := g.Clone()
	advanceIntoStep(t, g, StepDeclareBlockers)
	parked := g.Clone()
	if err := g.FinishBlocks(g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}

	g.RestoreFrom(parked)
	if g.Turn.Step != StepDeclareBlockers || g.Turn.PriorityHolder != NoPriority {
		t.Fatalf("undo to the parked step: %s, holder %d", g.Turn.Step, g.Turn.PriorityHolder)
	}
	if got := g.BlockDeclarationStatusOf(g.Seats[1].ID); got != BlockDeclarationPending {
		t.Fatalf("undo to the parked step: status %q, want pending", got)
	}
	g.RestoreFrom(beforeStep)
	if g.Turn.Step != StepDeclareAttackers || g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Fatalf("undo past the step: %s, holder %d", g.Turn.Step, g.Turn.PriorityHolder)
	}
	// And the table moves forward again from there.
	advanceIntoStep(t, g, StepDeclareBlockers)
	if g.Turn.PriorityHolder != NoPriority {
		t.Fatalf("re-entering the step parks priority again; holder %d", g.Turn.PriorityHolder)
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
