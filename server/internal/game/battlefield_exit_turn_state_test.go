package game

import (
	"testing"

	"github.com/google/uuid"
)

// battlefield_exit_turn_state_test.go — #630 / CR 400.7: the per-object
// state the ENGINE keeps does not survive a zone change any more than
// the state on the card does.
//
// The reported shape: activate Teferi's +1, return him to hand, recast
// him the same turn, and both loyalty rows are greyed out with
// "Already activated this turn" — because Game.LoyaltyActivatedThisTurn
// is keyed by instance ID, an instance ID survives battlefield → hand →
// stack → battlefield, and nothing deleted the entry until the turn
// ended.

// printedWalker is pushLoyaltyWalker plus the printed loyalty the
// battlefield entry stamps (CR 306.5b), so the walker can leave and
// come back with its own counters rather than none.
func printedWalker(t *testing.T, g *Game, owner *Player, loyalty, delta int) uuid.UUID {
	t.Helper()
	id := pushLoyaltyWalker(g, owner, loyalty, delta)
	c := findBattlefieldCard(g, id)
	if c == nil {
		t.Fatalf("test walker %s is not on the battlefield", id)
	}
	c.StartingLoyalty = loyalty
	return id
}

// TestLoyaltyAbilityIsAvailableAgainAfterTheWalkerLeavesAndReturns is
// the issue's repro. The permanent that comes back is a new object
// (CR 400.7) and CR 606.3 applies per object, so it gets its own
// activation.
func TestLoyaltyAbilityIsAvailableAgainAfterTheWalkerLeavesAndReturns(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pw := printedWalker(t, g, me, 4, 1)

	if err := g.ActivateCatalogAbility(me.ID, pw, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	_ = g.PassPriority()
	_ = g.PassPriority()
	if !g.LoyaltyActivatedThisTurn[pw] {
		t.Fatalf("setup: the activation did not register")
	}
	if got := loyaltyOf(g, pw); got != 5 {
		t.Fatalf("setup: loyalty after +1 = %d, want 5", got)
	}

	// Venser bounces it; the owner recasts it. The instance ID is the
	// same object identity the client and the replay see — which is
	// exactly why the stale entry was reachable.
	var back uuid.UUID
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(pw); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
		if g.LoyaltyActivatedThisTurn[pw] {
			t.Errorf("the activation followed the card into hand")
		}
		var err error
		back, err = g.PutFromHandOntoBattlefieldForEffect(pw, HandEntryOptions{Controller: me.ID})
		if err != nil {
			t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
		}
	})
	if back != pw {
		t.Fatalf("this path is meant to KEEP the instance ID (%s → %s); the repro depends on it", pw, back)
	}
	if g.Turn.ActiveSeat != 0 || g.Turn.Step != StepPrecombatMain {
		t.Fatalf("setup: left the turn (seat %d step %s)", g.Turn.ActiveSeat, g.Turn.Step)
	}
	if got := loyaltyOf(g, back); got != 4 {
		t.Errorf("the returning walker's loyalty = %d, want the printed 4", got)
	}

	if err := g.ActivateCatalogAbility(me.ID, back, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activation after the return: got %v, want it allowed (CR 400.7 / 606.3)", err)
	}
	if got := loyaltyOf(g, back); got != 5 {
		t.Errorf("loyalty after the new object's +1 = %d, want 5", got)
	}
}

// TestFlickeredPlaneswalkerGetsAFreshActivation — the blink path,
// which mints a new instance ID of its own (ADR 0026 §5). It already
// dodged the gate that way; what this pins is that the exit leaves no
// entry behind either, so the map does not accumulate the dead
// object's row for the rest of the turn.
func TestFlickeredPlaneswalkerGetsAFreshActivation(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pw := printedWalker(t, g, me, 4, 1)

	if err := g.ActivateCatalogAbility(me.ID, pw, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	_ = g.PassPriority()
	_ = g.PassPriority()

	var back uuid.UUID
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(pw); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
		var err error
		back, err = g.ReturnFromExileToBattlefieldForEffect(pw, me.ID, false)
		if err != nil {
			t.Fatalf("ReturnFromExileToBattlefieldForEffect: %v", err)
		}
	})
	if back == pw {
		t.Fatalf("the exile return is supposed to mint a new instance ID")
	}
	if n := len(g.LoyaltyActivatedThisTurn); n != 0 {
		t.Errorf("the flickered object left %d rows behind in LoyaltyActivatedThisTurn", n)
	}
	if got := loyaltyOf(g, back); got != 4 {
		t.Errorf("the flickered walker's loyalty = %d, want the printed 4", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, back, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activation after the flicker: got %v, want it allowed", err)
	}
}

// TestLoyaltyGateSurvivesAnotherPermanentLeaving — the forget is per
// OBJECT. A walker that stayed put keeps its spent activation even
// while something else leaves the battlefield, which is the half of
// CR 606.3 that must not be loosened.
func TestLoyaltyGateSurvivesAnotherPermanentLeaving(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	stays := printedWalker(t, g, me, 4, 1)
	leaves := printedWalker(t, g, me, 4, 1)
	// A different name, so the legend rule (CR 704.5j) does not hold
	// priority with a prompt and the +1 can resolve: the second
	// activation below must meet an EMPTY stack, or CR 307.1 refuses
	// it before CR 606.3 is asked (#1352).
	findBattlefieldCard(g, leaves).Name = "Another Test Planeswalker"

	if err := g.ActivateCatalogAbility(me.ID, stays, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activating the walker that stays: %v", err)
	}
	resolveWholeStackForTest(t, g)

	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(leaves); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
	})

	if !g.LoyaltyActivatedThisTurn[stays] {
		t.Fatalf("another permanent leaving cleared the standing walker's activation")
	}
	if err := g.ActivateCatalogAbility(me.ID, stays, 0, ActivateAbilityParams{}); err != ErrLoyaltyAlreadyActivated {
		t.Errorf("second activation without a zone change: got %v, want ErrLoyaltyAlreadyActivated", err)
	}
}

// TestUndoAcrossTheLoyaltyRefresh — the refresh is a map delete on
// Game, which Clone / RestoreFrom already carry (it is the same map
// the turn wrap clears). Rewinding to before the bounce puts the
// spent activation back.
func TestUndoAcrossTheLoyaltyRefresh(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pw := printedWalker(t, g, me, 4, 1)

	if err := g.ActivateCatalogAbility(me.ID, pw, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	_ = g.PassPriority()
	_ = g.PassPriority()
	snap := g.Clone()

	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(pw); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
		if _, err := g.PutFromHandOntoBattlefieldForEffect(pw, HandEntryOptions{Controller: me.ID}); err != nil {
			t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activation after the return: %v", err)
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })

	if !g.LoyaltyActivatedThisTurn[pw] {
		t.Errorf("undo did not put the spent activation back")
	}
	if findBattlefieldCard(g, pw) == nil {
		t.Fatalf("undo did not put the walker back on the battlefield")
	}
	if err := g.ActivateCatalogAbility(g.Seats[0].ID, pw, 0, ActivateAbilityParams{}); err != ErrLoyaltyAlreadyActivated {
		t.Errorf("activation after the undo: got %v, want ErrLoyaltyAlreadyActivated", err)
	}
}

// TestCombatAnnouncementsDoNotSurviveABattlefieldExit — the sibling
// per-object registries, cleared by the same exit. They are keyed by
// instance ID like the loyalty map and cleared only at end of combat,
// so a permanent that left mid-combat used to leave its row behind.
func TestCombatAnnouncementsDoNotSurviveABattlefieldExit(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	attacker := pushPlainCreature(t, g, g.Seats[0], "Alpha")
	blocker := pushPlainCreature(t, g, g.Seats[1], "Blocker")
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttackDeclaration(t, g)
	advanceIntoStep(t, g, StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	g.WithWriteLock(func() { g.commitBlockDeclarationLocked() })

	if !g.announcedAttacks[attacker] {
		t.Fatalf("setup: the attack was never announced")
	}
	if _, ok := g.announcedBlocks[blocker]; !ok {
		t.Fatalf("setup: the block was never announced")
	}
	if !g.blockedAttackers[attacker] {
		t.Fatalf("setup: the attacker was never recorded as blocked")
	}

	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(blocker); err != nil {
			t.Fatalf("BounceToHandForEffect(blocker): %v", err)
		}
		if err := g.BounceToHandForEffect(attacker); err != nil {
			t.Fatalf("BounceToHandForEffect(attacker): %v", err)
		}
	})

	if g.announcedAttacks[attacker] {
		t.Errorf("announcedAttacks kept a row for a permanent that left the battlefield")
	}
	if _, ok := g.announcedBlocks[blocker]; ok {
		t.Errorf("announcedBlocks kept a row for a permanent that left the battlefield")
	}
	if g.blockedAttackers[attacker] {
		t.Errorf("blockedAttackers kept a row for a permanent that left the battlefield")
	}
}

// TestSummoningSicknessMarkerDoesNotFollowTheCardOffTheBattlefield —
// the card-side sibling, in MoveCard's exit cleanup. "Entered the
// battlefield at" and the marker stamped with it belong to the
// permanent; every entry stamps its own pair, and a creature in a
// graveyard should not be telling the wire it is summoning sick.
func TestSummoningSicknessMarkerDoesNotFollowTheCardOffTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := pushScopedTestCreature(g, owner.ID, 2, 2)

	c := findBattlefieldCard(g, id)
	if c == nil || !c.SummonedThisTurn || c.EnteredBattlefieldAt == 0 {
		t.Fatalf("setup: the entry stamp did not land")
	}

	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(id); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
	})

	inHand := zoneCard(t, owner.Hand, id)
	if inHand.SummonedThisTurn {
		t.Errorf("the card in hand still claims it entered the battlefield this turn")
	}
	if inHand.EnteredBattlefieldAt != 0 {
		t.Errorf("the card in hand still carries its battlefield timestamp")
	}

	// And the permanent it becomes is stamped afresh.
	var back uuid.UUID
	g.WithWriteLock(func() {
		var err error
		back, err = g.PutFromHandOntoBattlefieldForEffect(id, HandEntryOptions{Controller: owner.ID})
		if err != nil {
			t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
		}
	})
	if c := findBattlefieldCard(g, back); c == nil || !c.SummonedThisTurn || c.EnteredBattlefieldAt == 0 {
		t.Errorf("the returning permanent was not stamped on entry")
	}
}
