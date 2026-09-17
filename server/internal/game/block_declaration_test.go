package game

import (
	"testing"

	"github.com/google/uuid"
)

// block_declaration_test.go covers #830 — the block declaration is
// announced at its LOCK-IN, not at each click.
//
// CR 509.1 declares blockers as one turn-based action and CR 506.4
// blocks each attacker once; the sandbox lets a defender re-point a
// blocker from one attacker to another before the declaration is
// complete. Announcing per click meant the attacker a blocker LEFT
// kept the "becomes blocked" trigger it had already fired — Cyberman
// Patrol's afflict landing on an attacker that ended up unblocked.

// lockInBlockDeclaration is the lock-in, called directly so a test
// can stay in the declare-blockers step across several of them. The
// two production paths into it (the priority wrap and the cursor
// leaving the step) have their own tests below.
func lockInBlockDeclaration(t *testing.T, g *Game) {
	t.Helper()
	g.WithWriteLock(func() { g.commitBlockDeclarationLocked() })
}

// blockDeclEvents returns the logged events of `kind` naming `card` in
// CardID — the blocker for EventBlock, the attacker for
// EventBecomesBlocked.
func blockDeclEvents(g *Game, kind EventKind, card uuid.UUID) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == kind && ev.CardID == card {
			out = append(out, ev)
		}
	}
	return out
}

// blockSetup is one attacking seat (0) with two attackers and one
// defending seat (1) with two untapped creatures, parked in the
// declare-blockers step.
func blockSetup(t *testing.T, g *Game) (attackerA, attackerB, blockerX, blockerY uuid.UUID) {
	t.Helper()
	attackerA = pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	attackerB = pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	blockerX = pushKeywordCreature(t, g, g.Seats[1], 1, 4)
	blockerY = pushKeywordCreature(t, g, g.Seats[1], 1, 4)
	advanceIntoStep(t, g, StepDeclareAttackers)
	for _, a := range []uuid.UUID{attackerA, attackerB} {
		if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	return attackerA, attackerB, blockerX, blockerY
}

// TestDeclareBlockerAnnouncesNothingBeforeTheLockIn is the half of
// #830 that makes the rest possible: a click stages the pairing and
// emits no event, so nothing has been harvested off it yet.
func TestDeclareBlockerAnnouncesNothingBeforeTheLockIn(t *testing.T) {
	g := newActiveGame(t)
	attackerA, _, blockerX, _ := blockSetup(t, g)

	if err := g.DeclareBlocker(blockerX, attackerA); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if n := len(blockDeclEvents(g, EventBlock, blockerX)); n != 0 {
		t.Fatalf("the click announces nothing: %d block events", n)
	}
	if n := len(blockDeclEvents(g, EventBecomesBlocked, attackerA)); n != 0 {
		t.Fatalf("the click announces nothing: %d becomes-blocked events", n)
	}
	if findCard(g, blockerX).BlockingTarget != attackerA {
		t.Error("the pairing is staged on the card all the same")
	}
}

// TestRepointedBlockerAnnouncesOnlyTheFinalAttacker is the #830
// reproduction: the attacker a blocker LEFT never became blocked.
func TestRepointedBlockerAnnouncesOnlyTheFinalAttacker(t *testing.T) {
	g := newActiveGame(t)
	attackerA, attackerB, blockerX, _ := blockSetup(t, g)

	if err := g.DeclareBlocker(blockerX, attackerA); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if err := g.DeclareBlocker(blockerX, attackerB); err != nil {
		t.Fatalf("re-point: %v", err)
	}
	lockInBlockDeclaration(t, g)

	blocks := blockDeclEvents(g, EventBlock, blockerX)
	if len(blocks) != 1 {
		t.Fatalf("one blocker blocks once (CR 509.3a): %d block events", len(blocks))
	}
	if blocks[0].Target != attackerB {
		t.Error("the block names the attacker the blocker ended on")
	}
	if n := len(blockDeclEvents(g, EventBecomesBlocked, attackerA)); n != 0 {
		t.Errorf("the attacker the blocker left is unblocked (CR 509.1h): %d becomes-blocked events", n)
	}
	if n := len(blockDeclEvents(g, EventBecomesBlocked, attackerB)); n != 1 {
		t.Errorf("the attacker it was re-pointed to becomes blocked once: %d events", n)
	}
}

// TestBlockerRepointedBackAnnouncesOnce — A, then B, then back to A.
// The round trip is one declaration, and A becomes blocked once.
func TestBlockerRepointedBackAnnouncesOnce(t *testing.T) {
	g := newActiveGame(t)
	attackerA, attackerB, blockerX, _ := blockSetup(t, g)

	for _, target := range []uuid.UUID{attackerA, attackerB, attackerA} {
		if err := g.DeclareBlocker(blockerX, target); err != nil {
			t.Fatalf("DeclareBlocker %s: %v", target, err)
		}
	}
	lockInBlockDeclaration(t, g)

	if n := len(blockDeclEvents(g, EventBlock, blockerX)); n != 1 {
		t.Errorf("three clicks are one block: %d block events", n)
	}
	if n := len(blockDeclEvents(g, EventBecomesBlocked, attackerA)); n != 1 {
		t.Errorf("the attacker it ended on becomes blocked once: %d events", n)
	}
	if n := len(blockDeclEvents(g, EventBecomesBlocked, attackerB)); n != 0 {
		t.Errorf("the attacker it passed through never became blocked: %d events", n)
	}
}

// TestDoubleBlockIsOneBecomesBlockedAndTwoBlocks — CR 506.4: an
// attacker is blocked once however many creatures block it, and each
// blocker blocks (CR 509.3a).
func TestDoubleBlockIsOneBecomesBlockedAndTwoBlocks(t *testing.T) {
	g := newActiveGame(t)
	attackerA, _, blockerX, blockerY := blockSetup(t, g)

	for _, b := range []uuid.UUID{blockerX, blockerY} {
		if err := g.DeclareBlocker(b, attackerA); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	lockInBlockDeclaration(t, g)

	if n := len(blockDeclEvents(g, EventBlock, blockerX)) + len(blockDeclEvents(g, EventBlock, blockerY)); n != 2 {
		t.Errorf("each blocker blocks: %d block events, want 2", n)
	}
	if n := len(blockDeclEvents(g, EventBecomesBlocked, attackerA)); n != 1 {
		t.Errorf("a double block is one 'becomes blocked' (CR 506.4): %d events", n)
	}
}

// TestBlockDeclarationIsOneEventBatch — the whole declaration is one
// occurrence, so a OncePerBatch ability collapses it (#854, CR
// 603.2c). Blockers are the mirror of DeclareAttacker, which relies
// on the same property for Adeline.
func TestBlockDeclarationIsOneEventBatch(t *testing.T) {
	g := newActiveGame(t)
	attackerA, attackerB, blockerX, blockerY := blockSetup(t, g)

	if err := g.DeclareBlocker(blockerX, attackerA); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if err := g.DeclareBlocker(blockerY, attackerB); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlockDeclaration(t, g)

	var batch uint64
	seen := 0
	for _, ev := range g.Events {
		if ev.Kind != EventBlock && ev.Kind != EventBecomesBlocked {
			continue
		}
		seen++
		if batch == 0 {
			batch = ev.Batch
			continue
		}
		if ev.Batch != batch {
			t.Fatalf("batch %d alongside %d: the declaration must be one batch", ev.Batch, batch)
		}
	}
	if seen != 4 {
		t.Fatalf("two blocks and two becomes-blocked: %d events", seen)
	}
}

// TestBlockDeclarationLocksInOnPriorityWrap — the production path
// through PassPriority. Priority passing all the way around is the
// declaration being complete (CR 509.1), and the engine announces it
// there rather than a step later.
func TestBlockDeclarationLocksInOnPriorityWrap(t *testing.T) {
	g := newActiveGame(t)
	attackerA, _, blockerX, _ := blockSetup(t, g)

	if err := g.DeclareBlocker(blockerX, attackerA); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	for i := 0; i < len(g.Seats); i++ {
		if g.Turn.Step != StepDeclareBlockers {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if n := len(blockDeclEvents(g, EventBecomesBlocked, attackerA)); n != 1 {
		t.Fatalf("the priority wrap locks the declaration in: %d becomes-blocked events", n)
	}
}

// TestBlockDeclarationLocksInBeforeTheCursorLeavesTheStep — the other
// production path. advance_step out of declare_blockers announces the
// declaration BEFORE the cursor moves, so the events belong to the
// step that produced them and their triggers are harvested there.
func TestBlockDeclarationLocksInBeforeTheCursorLeavesTheStep(t *testing.T) {
	g := newActiveGame(t)
	attackerA, _, blockerX, _ := blockSetup(t, g)

	if err := g.DeclareBlocker(blockerX, attackerA); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	becameBlocked := blockDeclEvents(g, EventBecomesBlocked, attackerA)
	if len(becameBlocked) != 1 {
		t.Fatalf("leaving the step locks the declaration in: %d becomes-blocked events", len(becameBlocked))
	}
	// The next step's announcement is later in the log, which is what
	// "before the cursor moved" means.
	for _, ev := range g.Events {
		if ev.Kind == EventStepBegan && ev.Step == StepCombatDamage && ev.Seq < becameBlocked[0].Seq {
			t.Error("the declaration was announced after the cursor had already entered combat damage")
		}
	}
}

// TestBlockAnnouncementsRewindWithUndo — both halves of an undo
// across a re-point. A clone taken while the declaration is staged
// comes back staged; a clone taken after the lock-in comes back with
// the announcements, so the same pairing is not announced twice.
func TestBlockAnnouncementsRewindWithUndo(t *testing.T) {
	g := newActiveGame(t)
	attackerA, attackerB, blockerX, _ := blockSetup(t, g)

	if err := g.DeclareBlocker(blockerX, attackerA); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	// Undo of a STAGED re-point: the blocker goes back to A, and the
	// lock-in then announces A — not B, which was never declared.
	staged := g.Clone()
	if err := g.DeclareBlocker(blockerX, attackerB); err != nil {
		t.Fatalf("re-point: %v", err)
	}
	g.WithWriteLock(func() { g.RestoreFrom(staged) })
	if findCard(g, blockerX).BlockingTarget != attackerA {
		t.Fatal("the undo takes the re-point back")
	}
	lockInBlockDeclaration(t, g)
	if n := len(blockDeclEvents(g, EventBecomesBlocked, attackerA)); n != 1 {
		t.Fatalf("one becomes-blocked for A after the undo: %d events", n)
	}
	if n := len(blockDeclEvents(g, EventBecomesBlocked, attackerB)); n != 0 {
		t.Fatalf("B was never declared against: %d events", n)
	}

	// Undo of a re-point made AFTER the lock-in. The announcements
	// rewind with the log, so re-locking announces nothing new.
	announced := g.Clone()
	if err := g.DeclareBlocker(blockerX, attackerB); err != nil {
		t.Fatalf("re-point: %v", err)
	}
	lockInBlockDeclaration(t, g)
	if n := len(blockDeclEvents(g, EventBecomesBlocked, attackerB)); n != 1 {
		t.Fatalf("the post-lock-in re-point announces B once: %d events", n)
	}
	g.WithWriteLock(func() { g.RestoreFrom(announced) })
	lockInBlockDeclaration(t, g)
	if n := len(blockDeclEvents(g, EventBlock, blockerX)); n != 1 {
		t.Errorf("the undone announcement is not replayed: %d block events", n)
	}
	if n := len(blockDeclEvents(g, EventBecomesBlocked, attackerA)); n != 1 {
		t.Errorf("A is still blocked exactly once after the undo: %d events", n)
	}
}

// TestClearCombatForgetsBlockAnnouncements — the announcements are
// one combat's bookkeeping. The same pairing next combat is a new
// declaration and announces again.
func TestClearCombatForgetsBlockAnnouncements(t *testing.T) {
	g := newActiveGame(t)
	attackerA, _, blockerX, _ := blockSetup(t, g)

	if err := g.DeclareBlocker(blockerX, attackerA); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlockDeclaration(t, g)
	if err := g.ClearCombat(); err != nil {
		t.Fatalf("ClearCombat: %v", err)
	}
	if err := g.DeclareBlocker(blockerX, attackerA); err != nil {
		t.Fatalf("DeclareBlocker again: %v", err)
	}
	lockInBlockDeclaration(t, g)
	if n := len(blockDeclEvents(g, EventBecomesBlocked, attackerA)); n != 2 {
		t.Errorf("a fresh combat announces afresh: %d becomes-blocked events, want 2", n)
	}
}
