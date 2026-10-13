package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// remove_from_combat_test.go — #672, CR 506.4: RemoveFromCombat is the
// sandbox verb and the "remove target creature from combat" effect's
// shared primitive. Regeneration (#667) and control changes reach the
// same removeFromCombatLocked; blocked_state_test.go pins the
// blocker-removed-attacker-stays-blocked half (CR 509.1h).

// A removed attacker hits nothing and its blockers, which stay
// blocking creatures, take and deal no combat damage.
func TestRemovedAttackerAndItsBlockersTakeNoCombatDamage(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	blocker := pushCombatant(t, g, g.Seats[1], "Blocker", 3, 3)
	blockAfterLockIn(t, g, attacker, blocker)

	if err := g.RemoveFromCombat(attacker); err != nil {
		t.Fatalf("RemoveFromCombat: %v", err)
	}
	if a := findCard(g, attacker); a == nil || a.AttackingTarget != uuid.Nil {
		t.Fatalf("the removed creature is still attacking (or left the battlefield)")
	}
	b := findCard(g, blocker)
	if b == nil || b.BlockingTarget != attacker {
		t.Fatalf("the blocker must stay a blocking creature after its attacker leaves combat")
	}

	seq := lastSeq(g)
	passUntilStep(t, g, StepCombatDamage)

	if got := StartingLife - g.Seats[1].Life; got != 0 {
		t.Errorf("defender lost %d life, want 0", got)
	}
	if got := combatDamageSince(g, seq); len(got) != 0 {
		t.Errorf("a combat with its only attacker removed dealt %d combat damage events: %+v", len(got), got)
	}
	if a := findCard(g, attacker); a.DamageMarked != 0 {
		t.Errorf("the removed attacker was dealt %d damage, want 0", a.DamageMarked)
	}
	if b := findCard(g, blocker); b.DamageMarked != 0 {
		t.Errorf("the blocker was dealt %d damage, want 0", b.DamageMarked)
	}
}

// A removed blocker deals no combat damage to the attacker and takes
// none; the attacker is still blocked, so it does not hit the player.
func TestRemovedBlockerDealsAndTakesNoCombatDamage(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	blocker := pushCombatant(t, g, g.Seats[1], "Blocker", 3, 3)
	blockAfterLockIn(t, g, attacker, blocker)

	if err := g.RemoveFromCombat(blocker); err != nil {
		t.Fatalf("RemoveFromCombat: %v", err)
	}
	if !g.blockedAttackers[attacker] {
		t.Fatalf("the attacker stopped being blocked (CR 509.1h)")
	}
	passUntilStep(t, g, StepCombatDamage)

	if got := StartingLife - g.Seats[1].Life; got != 0 {
		t.Errorf("defender lost %d life, want 0 — the attacker is still blocked", got)
	}
	if a := findCard(g, attacker); a.DamageMarked != 0 {
		t.Errorf("the attacker was dealt %d damage by a blocker removed from combat, want 0", a.DamageMarked)
	}
	if b := findCard(g, blocker); b.DamageMarked != 0 {
		t.Errorf("the removed blocker was dealt %d damage, want 0", b.DamageMarked)
	}
}

// Of two attackers, only the removed one is gone from combat: the
// other still connects.
func TestRemovingOneAttackerLeavesTheOtherHitting(t *testing.T) {
	g := newActiveGame(t)
	keep := pushCombatant(t, g, g.Seats[0], "Keeper", 3, 3)
	drop := pushCombatant(t, g, g.Seats[0], "Dropped", 2, 2)
	declareAttacks(t, g, keep, drop)

	if err := g.RemoveFromCombat(drop); err != nil {
		t.Fatalf("RemoveFromCombat: %v", err)
	}
	passUntilStep(t, g, StepCombatDamage)

	if got := StartingLife - g.Seats[1].Life; got != 3 {
		t.Errorf("defender lost %d life, want 3 (only the kept attacker)", got)
	}
}

func TestRemoveFromCombatRefusals(t *testing.T) {
	g := newActiveGame(t)
	idle := pushCombatant(t, g, g.Seats[0], "Idle", 2, 2)

	if err := g.RemoveFromCombat(idle); !errors.Is(err, ErrNotInCombat) {
		t.Errorf("a creature not in combat: got %v, want ErrNotInCombat", err)
	}
	if err := g.RemoveFromCombat(uuid.New()); !errors.Is(err, ErrCardNotFound) {
		t.Errorf("an unknown card: got %v, want ErrCardNotFound", err)
	}
}
