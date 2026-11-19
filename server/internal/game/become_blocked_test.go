package game

import (
	"testing"

	"github.com/google/uuid"
)

// become_blocked_test.go covers #5839: an effect that makes attacking
// creatures become blocked with no creature blocking them (Fog Patch,
// CR 509.1h). The attacker is blocked for the rest of combat, an
// unblocked one gets its one "becomes blocked" (CR 509.3c) and no
// EventBlock (CR 509.3d), and it assigns no combat damage unless it has
// trample, which assigns all of it to the player (CR 510.1c, 702.19d).

// becomeBlocked runs BecomeBlockedForEffect under the write lock.
func becomeBlocked(g *Game, ids ...uuid.UUID) int {
	var n int
	g.WithWriteLock(func() { n = g.BecomeBlockedForEffect(ids) })
	return n
}

func TestBecomeBlockedMakesAnUnblockedAttackerBlockedOnce(t *testing.T) {
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
	blocksBefore := 0
	for _, ev := range g.Events {
		if ev.Kind == EventBlock {
			blocksBefore++
		}
	}

	if n := becomeBlocked(g, attackerA, attackerB); n != 1 {
		t.Fatalf("BecomeBlockedForEffect marked %d attackers, want 1 (B was already blocked)", n)
	}

	evs := blockDeclEvents(g, EventBecomesBlocked, attackerA)
	if len(evs) != 1 {
		t.Fatalf("the unblocked attacker becomes blocked once (CR 509.3c): %d events", len(evs))
	}
	if ev := evs[0]; ev.Source != attackerA || ev.Target != attackerA || ev.Actor != defender {
		t.Errorf("the event has the lock-in's shape (attacker in Source/CardID/Target, defender as Actor): %+v", ev)
	}
	if n := len(blockDeclEvents(g, EventBecomesBlocked, attackerB)); n != 1 {
		t.Errorf("an already-blocked attacker does not become blocked again: %d events, want 1", n)
	}
	blocksAfter := 0
	for _, ev := range g.Events {
		if ev.Kind == EventBlock {
			blocksAfter++
		}
	}
	if blocksAfter != blocksBefore {
		t.Errorf("no creature blocks, so no block event (CR 509.3d): %d new", blocksAfter-blocksBefore)
	}

	var blocked, unblocked bool
	g.WithWriteLock(func() {
		blocked = g.BlockedAttackerForEffect(attackerA)
		unblocked = g.UnblockedAttackerForEffect(attackerA)
	})
	if !blocked || unblocked {
		t.Errorf("A reads blocked=%v unblocked=%v, want a blocked creature (CR 509.1h)", blocked, unblocked)
	}
	if got := findCard(g, attackerA).AttackingTarget; got != defender {
		t.Errorf("A is still attacking the defender: AttackingTarget %s", got)
	}

	// A second call changes nothing: it is blocked once.
	if n := becomeBlocked(g, attackerA); n != 0 {
		t.Errorf("a second call marked %d", n)
	}
	if n := len(blockDeclEvents(g, EventBecomesBlocked, attackerA)); n != 1 {
		t.Errorf("A became blocked %d times, want 1", n)
	}
}

// CR 510.1c and 702.19d: a blocked creature with no creature blocking it
// assigns no combat damage, unless it has trample, which assigns all of
// it to the player it is attacking.
func TestBecomeBlockedAttackerDealsNoDamageUnlessItTramples(t *testing.T) {
	g := newActiveGame(t)
	bears := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	trampler := pushCombatant(t, g, g.Seats[0], "Trampler", 3, 3, "trample")
	declareAttacks(t, g, bears, trampler)
	finishPendingBlockDeclarations(t, g)

	if n := becomeBlocked(g, bears, trampler); n != 2 {
		t.Fatalf("marked %d attackers, want 2", n)
	}
	seq := lastSeq(g)
	passUntilStep(t, g, StepCombatDamage)

	if got := StartingLife - g.Seats[1].Life; got != 3 {
		t.Errorf("defender lost %d life, want 3: the trampler's damage and none of the bears' (CR 510.1c, 702.19d)", got)
	}
	for _, ev := range combatDamageSince(g, seq) {
		if ev.Source == bears {
			t.Errorf("the blocked bears assigned combat damage: %+v", ev)
		}
	}
}

// The mark is Game.blockedAttackers, which rides the snapshot: a restore
// keeps the attacker blocked.
func TestBecomeBlockedSurvivesASnapshotRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	declareAttacks(t, g, attacker)
	finishPendingBlockDeclarations(t, g)
	becomeBlocked(g, attacker)

	_, restored := roundTrip(t, g)
	var blocked bool
	restored.WithWriteLock(func() { blocked = restored.BlockedAttackerForEffect(attacker) })
	if !blocked {
		t.Fatal("the restored attacker is no longer blocked")
	}
	passUntilStep(t, restored, StepCombatDamage)
	if got := StartingLife - restored.Seats[1].Life; got != 0 {
		t.Errorf("defender lost %d life after the restore, want 0 (CR 510.1c)", got)
	}
}

// Before its defending player has declared, an attacker is neither
// blocked nor unblocked, so it does not "become blocked" (CR 509.3c); it
// is marked blocked silently, as a creature put onto the battlefield
// blocking it marks it. Outside the combat steps from declare blockers
// on, nothing happens at all.
func TestBecomeBlockedBeforeTheDeclarationAndOutsideCombat(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	pushKeywordCreature(t, g, g.Seats[1], 1, 1) // a legal block keeps the declaration open
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if n := becomeBlocked(g, attacker); n != 0 {
		t.Errorf("in the declare-attackers step it marked %d", n)
	}

	advanceIntoStep(t, g, StepDeclareBlockers)
	if st := g.BlockDeclarationStatusOf(g.Seats[1].ID); st != BlockDeclarationPending {
		t.Fatalf("setup: the defender's declaration is %q, want pending", st)
	}
	if n := becomeBlocked(g, attacker); n != 1 {
		t.Fatalf("marked %d, want 1", n)
	}
	if n := len(blockDeclEvents(g, EventBecomesBlocked, attacker)); n != 0 {
		t.Errorf("not yet unblocked, so it does not become blocked: %d events", n)
	}
	var blocked bool
	g.WithWriteLock(func() { blocked = g.attackerBlockedLocked(attacker) })
	if !blocked {
		t.Error("the attacker is still marked blocked")
	}
}
