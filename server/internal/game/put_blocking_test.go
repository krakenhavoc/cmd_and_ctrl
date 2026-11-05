package game

import (
	"testing"

	"github.com/google/uuid"
)

// put_blocking_test.go covers #4956: a creature put onto the battlefield
// blocking an attacker that was UNBLOCKED at that time makes it "become
// blocked" (CR 509.3c), and one blocking an attacker already blocked
// does not make it become blocked again. Either way the entering
// creature was never declared as a blocker (CR 509.4), so it gets no
// EventBlock.

// createBlockingToken puts one 1/1 token onto the battlefield under
// controller's control, blocking attacker, and returns its ID.
func createBlockingToken(t *testing.T, g *Game, controller, attacker uuid.UUID) uuid.UUID {
	t.Helper()
	before := map[uuid.UUID]bool{}
	for _, c := range g.Battlefield.Cards {
		before[c.InstanceID] = true
	}
	g.WithWriteLock(func() {
		if err := g.CreateTokensBlockingForEffect(controller, Card{
			Name: "Cat Soldier", TypeLine: "Token Creature — Cat Soldier", Power: 1, Toughness: 1,
		}, 1, attacker); err != nil {
			t.Fatalf("CreateTokensBlockingForEffect: %v", err)
		}
	})
	for _, c := range g.Battlefield.Cards {
		if !before[c.InstanceID] {
			if c.BlockingTarget != attacker {
				t.Fatalf("the token is blocking %s, want %s", c.BlockingTarget, attacker)
			}
			return c.InstanceID
		}
	}
	t.Fatal("no token entered")
	return uuid.Nil
}

func TestEntryBlockerOnAnUnblockedAttackerMakesItBecomeBlocked(t *testing.T) {
	g := newActiveGame(t)
	attackerA, attackerB, blockerX, _ := blockSetup(t, g)
	defender := g.Seats[1].ID

	// B is blocked by a declared blocker; A is left unblocked.
	if err := g.DeclareBlocker(blockerX, attackerB); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if err := g.FinishBlocks(defender); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	if n := len(blockDeclEvents(g, EventBecomesBlocked, attackerA)); n != 0 {
		t.Fatalf("A is unblocked after the declaration: %d becomes-blocked events", n)
	}
	var unblocked bool
	g.WithWriteLock(func() { unblocked = g.UnblockedAttackerForEffect(attackerA) })
	if !unblocked {
		t.Fatal("A should read as an unblocked attacker before the token enters")
	}

	token := createBlockingToken(t, g, defender, attackerA)

	evs := blockDeclEvents(g, EventBecomesBlocked, attackerA)
	if len(evs) != 1 {
		t.Fatalf("a token put onto the battlefield blocking an unblocked attacker makes it become blocked once (CR 509.3c): %d events", len(evs))
	}
	if ev := evs[0]; ev.Source != attackerA || ev.Target != attackerA || ev.Actor != defender {
		t.Errorf("the event has the lock-in's shape (attacker in Source/CardID/Target, defender as Actor): %+v", ev)
	}
	if n := len(blockDeclEvents(g, EventBlock, token)); n != 0 {
		t.Errorf("the token was never declared as a blocker (CR 509.4): %d block events", n)
	}
	var blocked bool
	g.WithWriteLock(func() { blocked = g.attackerBlockedLocked(attackerA) })
	if !blocked {
		t.Error("A is now a blocked creature (CR 509.1h)")
	}

	// The next lock-in neither re-announces A nor announces the token.
	lockInBlockDeclaration(t, g)
	if n := len(blockDeclEvents(g, EventBecomesBlocked, attackerA)); n != 1 {
		t.Errorf("A becomes blocked once: %d events after the next lock-in", n)
	}
	if n := len(blockDeclEvents(g, EventBlock, token)); n != 0 {
		t.Errorf("the lock-in announced the token's block: %d events", n)
	}

	// B was already blocked: a second token blocking it does not make it
	// become blocked again.
	createBlockingToken(t, g, defender, attackerB)
	if n := len(blockDeclEvents(g, EventBecomesBlocked, attackerB)); n != 1 {
		t.Errorf("an already-blocked attacker does not become blocked again: %d events, want 1", n)
	}
}

// Before its defending player has declared blockers an attacker is
// neither blocked nor unblocked (CR 509.1h), so a creature put onto the
// battlefield blocking it does not make it "become blocked" — CR 509.3c
// asks for an attacker that was unblocked at that time.
func TestEntryBlockerBeforeTheDeclarationDoesNotMakeItBecomeBlocked(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}

	createBlockingToken(t, g, g.Seats[1].ID, attacker)

	if n := len(blockDeclEvents(g, EventBecomesBlocked, attacker)); n != 0 {
		t.Errorf("not yet unblocked, so it does not become blocked: %d events", n)
	}
	var blocked bool
	g.WithWriteLock(func() { blocked = g.attackerBlockedLocked(attacker) })
	if !blocked {
		t.Error("the attacker is still marked blocked")
	}
}
