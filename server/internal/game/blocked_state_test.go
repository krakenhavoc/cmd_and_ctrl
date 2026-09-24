package game

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// blocked_state_test.go — #715, CR 509.1h: "a creature remains blocked
// even if all the creatures blocking it are removed from combat".
//
// The engine used to have no blocked state at all. Each combat damage
// pass rebuilt the blocker map from the live battlefield, so an
// attacker whose blockers had died, been bounced or been removed from
// combat read as UNBLOCKED and hit the defending player, and a menace
// attacker that lost one of its two blockers had its block silently
// reverted at damage time.
//
// The fix is one record — Game.blockedAttackers — written in one place
// (commitBlockDeclarationLocked, the block declaration's lock-in) and
// read by the combat damage steps. These tests are the rules cases:
// CR 510.1c (blocked, no blockers left, no damage), CR 702.19d/e
// (trample is the exception) and CR 509.1b (block legality, menace
// included, is judged at the declaration and never again).

// blockAfterLockIn declares one attacker against seat 1, declares the
// given blockers against it AS ONE SET, and locks the declaration in
// the way a priority boundary inside declare_blockers does
// (CR 509.2a) — a trick cast in the step, or the cursor leaving it.
// Leaves the cursor in declare_blockers so the caller can remove a
// blocker before damage.
//
// One DeclareBlockers rather than a DeclareBlocker per creature: a
// count rule is judged on the whole declaration (#750), so a
// two-creature menace block only exists as a pair and declaring half
// of it is refused.
func blockAfterLockIn(t *testing.T, g *Game, attacker uuid.UUID, blockers ...uuid.UUID) {
	t.Helper()
	declareAttacks(t, g, attacker)
	decls := make([]BlockDeclaration, 0, len(blockers))
	for _, b := range blockers {
		decls = append(decls, BlockDeclaration{Blocker: b, Attacker: attacker})
	}
	if err := g.DeclareBlockers(decls); err != nil {
		t.Fatalf("DeclareBlockers: %v", err)
	}
	g.WithWriteLock(func() { g.completeAllBlockDeclarationsLocked(); g.commitBlockDeclarationLocked() })
	if !g.blockedAttackers[attacker] {
		t.Fatalf("setup: the attacker was not recorded as blocked by the lock-in")
	}
}

// CR 510.1c: a blocked creature with no creatures blocking it assigns
// no combat damage. The chump blocker dies to an instant in the
// declare-blockers window and the 2/2 hits nothing at all.
func TestBlockedAttackerWhoseOnlyBlockerDiedDealsNoDamage(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	chump := pushCombatant(t, g, g.Seats[1], "Chump", 1, 1)
	blockAfterLockIn(t, g, attacker, chump)

	seq := lastSeq(g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(chump); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	passUntilStep(t, g, StepCombatDamage)

	if got := StartingLife - g.Seats[1].Life; got != 0 {
		t.Errorf("defender lost %d life, want 0 — the attacker is still blocked (CR 509.1h)", got)
	}
	if got := combatDamageSince(g, seq); len(got) != 0 {
		t.Errorf("blocked attacker with no blockers left assigned %d combat damage events, want none (CR 510.1c): %+v",
			len(got), got)
	}
}

// A bounced blocker is the same rule by a different route (CR 506.4:
// a creature that leaves the battlefield is removed from combat), and
// it is the route that also drops the blocker's own announcement rows
// (#935) — the ATTACKER's blocked row is keyed by the attacker and
// must survive.
func TestBlockedAttackerWhoseOnlyBlockerWasBouncedDealsNoDamage(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	chump := pushCombatant(t, g, g.Seats[1], "Chump", 1, 1)
	blockAfterLockIn(t, g, attacker, chump)

	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(chump); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
	})
	if !g.blockedAttackers[attacker] {
		t.Fatalf("the attacker stopped being blocked when its blocker left the battlefield")
	}
	passUntilStep(t, g, StepCombatDamage)

	if got := StartingLife - g.Seats[1].Life; got != 0 {
		t.Errorf("defender lost %d life, want 0 (CR 509.1h)", got)
	}
}

// CR 702.19d/e: trample is the exception. A blocked trampler with no
// creatures blocking it assigns all its damage to the player or
// planeswalker it is attacking.
func TestBlockedTramplerWhoseBlockerDiedHitsThePlayer(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Trampler", 3, 3, "trample")
	chump := pushCombatant(t, g, g.Seats[1], "Chump", 1, 1)
	blockAfterLockIn(t, g, attacker, chump)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(chump); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	passUntilStep(t, g, StepCombatDamage)

	if got := StartingLife - g.Seats[1].Life; got != 3 {
		t.Errorf("defender lost %d life, want 3 — a blocked trampler with no blockers tramples over (CR 702.19d/e)", got)
	}
}

// A blocker taken out of combat by an EFFECT (#921's
// removeFromCombatLocked, the route a control change and the #672 verb
// take) leaves the attacker blocked too: it is still on the
// battlefield, it is simply no longer blocking.
func TestBlockerRemovedFromCombatLeavesTheAttackerBlocked(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	blocker := pushCombatant(t, g, g.Seats[1], "Blocker", 1, 4)
	blockAfterLockIn(t, g, attacker, blocker)

	g.WithWriteLock(func() { g.removeFromCombatLocked(findBattlefieldCard(g, blocker)) })
	if !g.blockedAttackers[attacker] {
		t.Fatalf("the attacker stopped being blocked when its blocker was removed from combat")
	}
	passUntilStep(t, g, StepCombatDamage)

	if got := StartingLife - g.Seats[1].Life; got != 0 {
		t.Errorf("defender lost %d life, want 0 (CR 509.1h)", got)
	}
	if b := findCard(g, blocker); b == nil {
		t.Fatalf("the blocker left the battlefield; it was only removed from combat")
	} else if b.DamageMarked != 0 {
		t.Errorf("a creature removed from combat was dealt %d combat damage, want 0", b.DamageMarked)
	}
}

// Menace (CR 702.111b) is a block-COUNT rule, and CR 509.1b judges it
// on the declaration. Two blockers were declared, so the block is
// legal; one of them dying before damage does not revert it (#715 —
// the old close-out ran on the live map at damage time and handed the
// defender three damage they had blocked).
func TestMenaceBlockSurvivesOneBlockerLeaving(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Menacer", 3, 3, "menace")
	first := pushCombatant(t, g, g.Seats[1], "Blocker One", 1, 1)
	second := pushCombatant(t, g, g.Seats[1], "Blocker Two", 1, 5)
	blockAfterLockIn(t, g, attacker, first, second)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(first); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	passUntilStep(t, g, StepCombatDamage)

	if got := StartingLife - g.Seats[1].Life; got != 0 {
		t.Errorf("defender lost %d life, want 0 — the menace block stands (CR 509.1h)", got)
	}
	blk := findCard(g, second)
	if blk == nil {
		t.Fatalf("the remaining blocker is gone")
	}
	if blk.DamageMarked != 3 {
		t.Errorf("the remaining blocker was dealt %d, want 3 — all of the attacker's damage", blk.DamageMarked)
	}
}

// The other half of the same rule: an illegal count IS refused, and
// #750 moves that refusal to the DECLARATION. The lone block is never
// stored, so there is nothing to revert at the lock-in and nothing
// the defender was told was good; the menace attacker is unblocked
// because the block never happened.
func TestMenaceRefusesASingleBlockerAtDeclaration(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Menacer", 3, 3, "menace")
	lone := pushCombatant(t, g, g.Seats[1], "Lone Blocker", 2, 2)

	declareAttacks(t, g, attacker)
	err := g.DeclareBlocker(lone, attacker)
	if !errors.Is(err, ErrIllegalBlock) {
		t.Fatalf("DeclareBlocker on a menace attacker = %v, want an illegal-block refusal", err)
	}
	var refusal *BlockRefusedError
	if !errors.As(err, &refusal) {
		t.Fatalf("the refusal carries no reason: %v", err)
	}
	if refusal.Reason != BlockReasonTooFewBlockers || refusal.N != 2 {
		t.Errorf("refusal = %q with N = %d, want %q with 2", refusal.Reason, refusal.N, BlockReasonTooFewBlockers)
	}
	if got := refusal.Sentence(g.Seats[1].ID); got != "Menacer can't be blocked by fewer than two creatures." {
		t.Errorf("the defender reads %q", got)
	}

	if g.Turn.Step != StepDeclareBlockers {
		t.Fatalf("setup: the cursor left declare_blockers, at %q", g.Turn.Step)
	}
	if c := findCard(g, lone); c == nil || c.BlockingTarget != uuid.Nil {
		t.Errorf("a refused block was stored anyway (CR 509.1b)")
	}
	if g.blockedAttackers[attacker] {
		t.Errorf("a menace attacker with one blocker was recorded as blocked")
	}

	passUntilStep(t, g, StepCombatDamage)
	if got := StartingLife - g.Seats[1].Life; got != 3 {
		t.Errorf("defender lost %d life, want 3 — the attacker ended up unblocked", got)
	}
}

// An undo that rewinds to before the removal keeps the blocked state,
// so the replay reaches the same board. A restore that dropped it
// would hand the defender damage they had blocked.
func TestUndoAcrossTheBlockerRemovalKeepsTheBlockedState(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	chump := pushCombatant(t, g, g.Seats[1], "Chump", 1, 1)
	blockAfterLockIn(t, g, attacker, chump)

	blocked := g.Clone()

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(chump); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	passUntilStep(t, g, StepCombatDamage)
	if got := StartingLife - g.Seats[1].Life; got != 0 {
		t.Fatalf("defender lost %d life before the undo, want 0", got)
	}

	g.WithWriteLock(func() { g.RestoreFrom(blocked) })
	if g.Turn.Step != StepDeclareBlockers {
		t.Fatalf("after the undo the cursor is at %q, want %q", g.Turn.Step, StepDeclareBlockers)
	}
	if !g.blockedAttackers[attacker] {
		t.Fatalf("the undo dropped the blocked state")
	}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(chump); err != nil {
			t.Fatalf("DestroyPermanentForEffect after the undo: %v", err)
		}
	})
	passUntilStep(t, g, StepCombatDamage)
	if got := StartingLife - g.Seats[1].Life; got != 0 {
		t.Errorf("defender lost %d life on the replay, want 0", got)
	}
}

// A game saved mid-combat with the blockers already gone comes back
// with the attacker still blocked, and deals the same damage — none.
func TestSnapshotRoundTripKeepsTheBlockedState(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	chump := pushCombatant(t, g, g.Seats[1], "Chump", 1, 1)
	blockAfterLockIn(t, g, attacker, chump)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(chump); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}

	if !restored.blockedAttackers[attacker] {
		t.Fatalf("the restored game lost the blocked state")
	}
	passUntilStep(t, restored, StepCombatDamage)
	if got := StartingLife - restored.Seats[1].Life; got != 0 {
		t.Errorf("the restored defender lost %d life, want 0", got)
	}
}
