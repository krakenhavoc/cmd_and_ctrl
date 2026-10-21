package game

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// loyalty_extra_test.go — #4191: The Chain Veil's "for each planeswalker
// you control, you may activate one of its loyalty abilities once this
// turn as though none of its loyalty abilities have been activated this
// turn" (GrantExtraLoyaltyActivationForEffect; ADR 0145 amendment
// 2026-10-21).

// activateLoyaltyAndResolve activates the walker's one loyalty row and
// lets it resolve.
func activateLoyaltyAndResolve(t *testing.T, g *Game, p *Player, pw uuid.UUID) error {
	t.Helper()
	if err := g.ActivateCatalogAbility(p.ID, pw, 0, ActivateAbilityParams{}); err != nil {
		return err
	}
	_ = g.PassPriority()
	_ = g.PassPriority()
	return nil
}

func grantExtraLoyalty(g *Game, p *Player) {
	g.WithWriteLock(func() { g.GrantExtraLoyaltyActivationForEffect(p.ID) })
}

func loyaltySpent(g *Game, pw uuid.UUID) bool {
	var spent bool
	g.ReadSnapshot(func() { spent = g.LoyaltySpentLocked(pw) })
	return spent
}

// After one activation the grant opens exactly one more, and the gate,
// the shared predicate and the second activation's effect agree.
func TestExtraLoyaltyActivationOpensOneMoreAfterTheFirst(t *testing.T) {
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
	grantExtraLoyalty(g, me)
	if loyaltySpent(g, pw) {
		t.Fatal("the grant did not open another activation")
	}
	if err := activateLoyaltyAndResolve(t, g, me, pw); err != nil {
		t.Fatalf("second activation under the grant: %v", err)
	}
	if got := counterOf(g, pw, "effect-ran"); got != 2 {
		t.Errorf("effect ran %d times, want 2", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, ActivateAbilityParams{}); !errors.Is(err, ErrLoyaltyAlreadyActivated) {
		t.Errorf("third activation: got %v, want ErrLoyaltyAlreadyActivated", err)
	}
	if got := loyaltyOf(g, pw); got != 6 {
		t.Errorf("loyalty %d, want 6 after two +1s", got)
	}
}

// A grant made before the walker has activated gives it two in all
// (the ruling: "a total of twice during your turn").
func TestExtraLoyaltyActivationBeforeAnyActivationMakesTwo(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pw := pushLoyaltyWalker(g, me, 4, 1)
	grantExtraLoyalty(g, me)

	for i := 1; i <= 2; i++ {
		if err := activateLoyaltyAndResolve(t, g, me, pw); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
	}
	if !loyaltySpent(g, pw) {
		t.Error("a third activation is open after one grant")
	}
}

// The grant changes the rules for the player, so a planeswalker that
// comes under their control after it resolves has it too.
func TestExtraLoyaltyActivationCoversAWalkerThatArrivesLater(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	grantExtraLoyalty(g, me)
	pw := pushLoyaltyWalker(g, me, 4, 1)

	for i := 1; i <= 2; i++ {
		if err := activateLoyaltyAndResolve(t, g, me, pw); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
	}
	if !loyaltySpent(g, pw) {
		t.Error("a third activation is open after one grant")
	}
}

// Two grants (an untapped Veil activated again) make three activations.
func TestExtraLoyaltyActivationGrantsStack(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pw := pushLoyaltyWalker(g, me, 4, 1)
	grantExtraLoyalty(g, me)
	grantExtraLoyalty(g, me)

	for i := 1; i <= 3; i++ {
		if err := activateLoyaltyAndResolve(t, g, me, pw); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
	}
	if !loyaltySpent(g, pw) {
		t.Error("a fourth activation is open after two grants")
	}
	if got := g.LoyaltyExtraActivationsUsedThisTurn[pw][me.ID]; got != 2 {
		t.Errorf("grants spent %d, want 2", got)
	}
}

// Another player's grant does nothing for your planeswalkers, and the
// grant is not a planeswalker's for a permanent that is not one.
func TestExtraLoyaltyActivationIsOnlyTheGranteesPlaneswalkers(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	pw := pushLoyaltyWalker(g, me, 4, 1)
	grantExtraLoyalty(g, opp)

	if err := activateLoyaltyAndResolve(t, g, me, pw); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	if !loyaltySpent(g, pw) {
		t.Error("an opponent's grant opened a second activation of your walker")
	}

	// A non-planeswalker with a loyalty row (CR 606.2) keeps one.
	grantExtraLoyalty(g, me)
	if loyaltySpent(g, pw) {
		t.Fatal("setup: your own grant did not open the walker")
	}
	artifact := NewCard("Loyal Artifact", me.ID)
	artifact.TypeLine = "Artifact"
	artifact.Controller = me.ID
	artifact.Counters = map[string]int{"loyalty": 3}
	artifact.ActivatedAbilities = []ActivatedAbilityShape{{
		Label:  "loyalty ability",
		Cost:   AbilityCost{Loyalty: loyaltyN(1)},
		Effect: func(*Game, *StackItem) error { return nil },
	}}
	g.Battlefield.PushTop(artifact)
	if err := activateLoyaltyAndResolve(t, g, me, artifact.InstanceID); err != nil {
		t.Fatalf("the artifact's first activation: %v", err)
	}
	if !loyaltySpent(g, artifact.InstanceID) {
		t.Error("the grant opened a second activation of a permanent that is not a planeswalker")
	}
}

// The sandbox loyalty verb asks the same predicate and spends the same
// grant.
func TestExtraLoyaltyActivationThroughTheSandboxVerb(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pw := pushLoyaltyWalker(g, me, 4, 1)
	grantExtraLoyalty(g, me)

	if err := g.ActivateLoyalty(me.ID, pw, "+1", 1); err != nil {
		t.Fatalf("first sandbox activation: %v", err)
	}
	if err := g.ActivateLoyalty(me.ID, pw, "+1", 1); err != nil {
		t.Fatalf("second sandbox activation under the grant: %v", err)
	}
	if err := g.ActivateLoyalty(me.ID, pw, "+1", 1); !errors.Is(err, ErrLoyaltyAlreadyActivated) {
		t.Errorf("third sandbox activation: got %v, want ErrLoyaltyAlreadyActivated", err)
	}
}

// The grants and what each permanent spent survive a snapshot round
// trip, and the turn boundary flushes both.
func TestExtraLoyaltyActivationSurvivesASnapshotAndEndsWithTheTurn(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pw := pushLoyaltyWalker(g, me, 4, 1)
	grantExtraLoyalty(g, me)
	grantExtraLoyalty(g, me)
	for i := 1; i <= 2; i++ {
		if err := activateLoyaltyAndResolve(t, g, me, pw); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
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
	if got := restored.LoyaltyExtraActivationsThisTurn[me.ID]; got != 2 {
		t.Errorf("restored grants %d, want 2", got)
	}
	if got := restored.LoyaltyExtraActivationsUsedThisTurn[pw][me.ID]; got != 1 {
		t.Errorf("restored grants spent %d, want 1", got)
	}
	if loyaltySpent(restored, pw) {
		t.Error("the restored walker lost its third activation")
	}

	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if g.LoyaltyExtraActivationsThisTurn != nil || g.LoyaltyExtraActivationsUsedThisTurn != nil {
		t.Errorf("the turn boundary kept %v / %v", g.LoyaltyExtraActivationsThisTurn, g.LoyaltyExtraActivationsUsedThisTurn)
	}
}

// A walker that leaves and comes back is a new object (CR 400.7): what
// the old one spent is forgotten, and the player's grant still covers
// the new one.
func TestExtraLoyaltyActivationSpentIsForgottenWhenTheWalkerLeaves(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pw := printedWalker(t, g, me, 4, 1)
	grantExtraLoyalty(g, me)
	for i := 1; i <= 2; i++ {
		if err := activateLoyaltyAndResolve(t, g, me, pw); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
	}
	var back uuid.UUID
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(pw); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
		if _, ok := g.LoyaltyExtraActivationsUsedThisTurn[pw]; ok {
			t.Error("what the walker spent followed the card into hand")
		}
		var err error
		back, err = g.PutFromHandOntoBattlefieldForEffect(pw, HandEntryOptions{Controller: me.ID})
		if err != nil {
			t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
		}
	})
	for i := 1; i <= 2; i++ {
		if err := activateLoyaltyAndResolve(t, g, me, back); err != nil {
			t.Fatalf("the returned walker's activation %d: %v", i, err)
		}
	}
	if !loyaltySpent(g, back) {
		t.Error("the returned walker has a third activation")
	}
}
