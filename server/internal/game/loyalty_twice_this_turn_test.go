package game

import (
	"encoding/json"
	"errors"
	"testing"
)

// loyalty_twice_this_turn_test.go — #5665: "You may activate the
// loyalty abilities of planeswalkers you control twice this turn rather
// than only once" (AllowLoyaltyTwiceThisTurnForEffect; ADR 0145
// amendment 2026-11-12).

func allowLoyaltyTwice(g *Game, p *Player) {
	g.WithWriteLock(func() { g.AllowLoyaltyTwiceThisTurnForEffect(p.ID) })
}

// The effect makes the limit two, an activation made before it counts
// as the first, and a third is refused.
func TestLoyaltyTwiceThisTurnMakesTheLimitTwo(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pw := pushLoyaltyWalker(g, me, 4, 1)

	if err := activateLoyaltyAndResolve(t, g, me, pw); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	if !loyaltySpent(g, pw) {
		t.Fatal("setup: the first activation did not use up the turn")
	}
	allowLoyaltyTwice(g, me)
	if loyaltySpent(g, pw) {
		t.Fatal("the effect did not open a second activation")
	}
	if err := activateLoyaltyAndResolve(t, g, me, pw); err != nil {
		t.Fatalf("second activation: %v", err)
	}
	if got := counterOf(g, pw, "effect-ran"); got != 2 {
		t.Fatalf("the ability resolved %d times, want 2", got)
	}
	if !loyaltySpent(g, pw) {
		t.Error("two activations did not use up the turn")
	}
	if err := activateLoyaltyAndResolve(t, g, me, pw); !errors.Is(err, ErrLoyaltyAlreadyActivated) {
		t.Errorf("third activation: got %v, want ErrLoyaltyAlreadyActivated", err)
	}
}

// Twice, not once more: a second resolution of the effect leaves the
// limit at two.
func TestLoyaltyTwiceThisTurnDoesNotStackWithItself(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pw := pushLoyaltyWalker(g, me, 4, 1)
	allowLoyaltyTwice(g, me)
	allowLoyaltyTwice(g, me)

	for i := 1; i <= 2; i++ {
		if err := activateLoyaltyAndResolve(t, g, me, pw); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
	}
	if !loyaltySpent(g, pw) {
		t.Fatal("two resolutions of the effect opened a third activation")
	}
}

// It covers a planeswalker that comes under the player's control after
// the effect resolved, and stacks with a Chain Veil grant, which is one
// activation on top of the limit.
func TestLoyaltyTwiceThisTurnCoversALaterWalkerAndStacksWithAGrant(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	allowLoyaltyTwice(g, me)
	grantExtraLoyalty(g, me)
	pw := pushLoyaltyWalker(g, me, 4, 1)

	for i := 1; i <= 3; i++ {
		if loyaltySpent(g, pw) {
			t.Fatalf("no activation %d was open", i)
		}
		if err := activateLoyaltyAndResolve(t, g, me, pw); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
	}
	if !loyaltySpent(g, pw) {
		t.Error("twice plus one grant opened a fourth activation")
	}
}

// The effect is the player's: another player's planeswalker keeps the
// limit of one.
func TestLoyaltyTwiceThisTurnIsOnlyThatPlayers(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, other := g.Seats[0], g.Seats[1]
	pw := pushLoyaltyWalker(g, me, 4, 1)
	allowLoyaltyTwice(g, other)

	if err := activateLoyaltyAndResolve(t, g, me, pw); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	if !loyaltySpent(g, pw) {
		t.Error("another player's effect opened a second activation")
	}
}

// The effect survives a snapshot round trip and ends with the turn.
func TestLoyaltyTwiceThisTurnSurvivesASnapshotAndEndsWithTheTurn(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pw := pushLoyaltyWalker(g, me, 4, 1)
	allowLoyaltyTwice(g, me)
	if err := activateLoyaltyAndResolve(t, g, me, pw); err != nil {
		t.Fatalf("first activation: %v", err)
	}

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if !restored.LoyaltyTwiceThisTurn[me.ID] {
		t.Error("the restored game lost the effect")
	}
	if loyaltySpent(restored, pw) {
		t.Error("the restored walker lost its second activation")
	}

	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if g.LoyaltyTwiceThisTurn != nil {
		t.Errorf("the turn boundary kept %v", g.LoyaltyTwiceThisTurn)
	}
}
