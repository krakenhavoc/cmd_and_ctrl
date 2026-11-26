package game

import (
	"testing"

	"github.com/google/uuid"
)

// combat_reassign_test.go — #6193: blocks changed by an effect after
// blockers are declared (CR 509.1h, 509.3). Balduvian Warlord, Ydwen
// Efreet and False Orders remove a blocker and unblock what only it had
// blocked, then make it block another attacker; Sorrow's Path swaps two
// blockers; General Jarkeld swaps two attackers' blockers.

// reassignSetup declares `attackers` at seat 1, has seat 1 block per
// `blocks` (blocker, attacker, ...), and finishes the declaration.
func reassignSetup(t *testing.T, g *Game, attackers []uuid.UUID, blocks ...uuid.UUID) {
	t.Helper()
	declareAttacks(t, g, attackers...)
	if len(blocks) > 0 {
		if err := block(t, g, blocks...); err != nil {
			t.Fatalf("blocks: %v", err)
		}
	}
	if err := g.FinishBlocks(g.Seats[1].ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
}

func blockedNow(g *Game, id uuid.UUID) bool {
	var out bool
	g.WithWriteLock(func() { out = g.attackerBlockedLocked(id) })
	return out
}

func blockingOf(g *Game, id uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	g.WithWriteLock(func() { out = findBattlefieldCard(g, id).BlockedAttackers() })
	return out
}

// Balduvian Warlord: the blocker leaves combat, its attacker becomes
// unblocked and deals its damage to the player, and the blocker then
// blocks another attacker, which becomes blocked once.
func TestRemovedBlockerUnblocksItsAttackerAndBlocksAnother(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := pushCombatant(t, g, me, "First Bear", 3, 3)
	second := pushCombatant(t, g, me, "Second Bear", 2, 2)
	wall := pushCombatant(t, g, opp, "Wall", 0, 5)
	reassignSetup(t, g, []uuid.UUID{first, second}, wall, first)
	if !blockedNow(g, first) || blockedNow(g, second) {
		t.Fatal("setup: the first attacker is not the blocked one")
	}
	becameBlocked := eventsOfKind(g, EventBecomesBlocked)
	var unblocked []uuid.UUID
	var n int
	g.WithWriteLock(func() {
		unblocked = g.RemoveBlockerUnblockingForEffect(wall)
		n = g.BlockAttackersForEffect(wall, []uuid.UUID{second})
	})
	if len(unblocked) != 1 || unblocked[0] != first || blockedNow(g, first) {
		t.Fatalf("unblocked %v; the first attacker blocked = %v", unblocked, blockedNow(g, first))
	}
	if n != 1 || !blockedNow(g, second) {
		t.Fatalf("the wall blocks %d new attackers; second blocked = %v", n, blockedNow(g, second))
	}
	if got := eventsOfKind(g, EventBecomesBlocked) - becameBlocked; got != 1 {
		t.Errorf("%d EventBecomesBlocked for the reassigned block, want 1 (CR 509.3c)", got)
	}
	life := opp.Life
	passUntilStep(t, g, StepEndCombat)
	if opp.Life != life-3 {
		t.Errorf("defender's life %d -> %d, want the unblocked attacker's 3 damage", life, opp.Life)
	}
}

// An attacker another creature had blocked this combat stays blocked,
// even when that other creature has already left combat.
func TestRemovedBlockerLeavesAnAttackerAnotherCreatureBlocked(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := pushCombatant(t, g, me, "Bear", 3, 3)
	w1 := pushCombatant(t, g, opp, "Wall One", 0, 5)
	w2 := pushCombatant(t, g, opp, "Wall Two", 0, 5)
	reassignSetup(t, g, []uuid.UUID{atk}, w1, atk, w2, atk)
	g.WithWriteLock(func() {
		if err := g.RemoveFromCombatForEffect(w1); err != nil {
			t.Fatal(err)
		}
	})
	var unblocked []uuid.UUID
	g.WithWriteLock(func() { unblocked = g.RemoveBlockerUnblockingForEffect(w2) })
	if len(unblocked) != 0 || !blockedNow(g, atk) {
		t.Errorf("unblocked %v: the attacker had become blocked by another creature this combat", unblocked)
	}
}

// Sorrow's Path: two blockers swap; one that couldn't block the other's
// attacker (flying) refuses the swap and nothing changes.
func TestSwapBlockersForEffect(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a1 := pushCombatant(t, g, me, "Bear", 2, 2)
	a2 := pushCombatant(t, g, me, "Ogre", 3, 3)
	flier := pushCombatant(t, g, me, "Bird", 1, 1, "flying")
	b1 := pushCombatant(t, g, opp, "Wall One", 0, 5)
	b2 := pushCombatant(t, g, opp, "Wall Two", 0, 5, "reach")
	reassignSetup(t, g, []uuid.UUID{a1, a2, flier}, b1, a1, b2, flier)
	var ok bool
	g.WithWriteLock(func() { ok = g.SwapBlockersForEffect(b1, b2) })
	if ok {
		t.Fatal("a wall without flying or reach was swapped onto the flier")
	}
	if got := blockingOf(g, b1); len(got) != 1 || got[0] != a1 {
		t.Errorf("after the refused swap Wall One blocks %v", got)
	}
	// Re-point Wall Two onto the Ogre with a declaration-free effect,
	// then the swap is legal both ways.
	g.WithWriteLock(func() {
		g.removeFromCombatLocked(findBattlefieldCard(g, b2))
		g.BlockAttackersForEffect(b2, []uuid.UUID{a2})
	})
	g.WithWriteLock(func() { ok = g.SwapBlockersForEffect(b1, b2) })
	if !ok {
		t.Fatal("a legal swap was refused")
	}
	if got := blockingOf(g, b1); len(got) != 1 || got[0] != a2 {
		t.Errorf("Wall One blocks %v, want the Ogre", got)
	}
	if got := blockingOf(g, b2); len(got) != 1 || got[0] != a1 {
		t.Errorf("Wall Two blocks %v, want the Bear", got)
	}
	if !blockedNow(g, a1) || !blockedNow(g, a2) {
		t.Error("an attacker stopped being blocked across the swap")
	}
}

// General Jarkeld: the attackers' blockers swap; a creature blocking both
// stays as it is, and no attacker becomes blocked again.
func TestSwitchBlockedAttackersForEffect(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	x := pushCombatant(t, g, me, "Bear", 2, 2)
	y := pushCombatant(t, g, me, "Ogre", 3, 3)
	bx := pushCombatant(t, g, opp, "Wall One", 0, 5)
	by := pushCombatant(t, g, opp, "Wall Two", 0, 5)
	reassignSetup(t, g, []uuid.UUID{x, y}, bx, x, by, y)
	became := eventsOfKind(g, EventBecomesBlocked)
	var ok bool
	g.WithWriteLock(func() { ok = g.SwitchBlockedAttackersForEffect(x, y) })
	if !ok {
		t.Fatal("the switch was refused")
	}
	if got := blockingOf(g, bx); len(got) != 1 || got[0] != y {
		t.Errorf("Wall One blocks %v, want the Ogre", got)
	}
	if got := blockingOf(g, by); len(got) != 1 || got[0] != x {
		t.Errorf("Wall Two blocks %v, want the Bear", got)
	}
	if eventsOfKind(g, EventBecomesBlocked) != became {
		t.Error("a switched attacker became blocked again (#830)")
	}
}

// Outside the declare-blockers step and the steps after it, the effects
// do nothing.
func TestReassignmentOnlyAfterBlockersAreDeclared(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := pushCombatant(t, g, me, "Bear", 2, 2)
	wall := pushCombatant(t, g, opp, "Wall", 0, 5)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, opp.ID); err != nil {
		t.Fatal(err)
	}
	var n int
	g.WithWriteLock(func() { n = g.BlockAttackersForEffect(wall, []uuid.UUID{atk}) })
	if n != 0 {
		t.Error("an effect made a creature block during declare attackers")
	}
}
