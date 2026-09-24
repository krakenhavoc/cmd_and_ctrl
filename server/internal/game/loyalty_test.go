package game

import (
	"testing"

	"github.com/google/uuid"
)

// loyalty_test.go — issues #329 and #334: "I cast Teferi and when I
// click on him to choose one of his abilities it just tapped him" /
// "Planeswalker loyalty abilities are not able to be activated."
//
// Both replays show the same engine-side shape: the planeswalker is
// on the battlefield with the right loyalty counters (ADR 0032 §1
// works) and `activated_abilities: []` — there was no loyalty
// ability for the client to offer, because AbilityCost had no
// loyalty component. These tests pin the component, its payment,
// CR 606.6 and CR 606.3.

// loyaltyN is the pointer form a catalog card writes its loyalty
// cost in. [0] is a real printed cost, so the zero value has to be
// distinguishable from "no loyalty component".
func loyaltyN(n int) *int { return &n }

// pushLoyaltyWalker seats a planeswalker with `loyalty` counters and
// one loyalty ability costing `delta`, whose effect bumps a
// test-visible counter on the walker itself.
func pushLoyaltyWalker(g *Game, owner *Player, loyalty, delta int) uuid.UUID {
	c := NewCard("Test Planeswalker", owner.ID)
	c.TypeLine = "Legendary Planeswalker — Test"
	c.Controller = owner.ID
	c.Counters = map[string]int{"loyalty": loyalty}
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "loyalty ability",
		Cost:  AbilityCost{Loyalty: loyaltyN(delta)},
		Effect: func(g *Game, item *StackItem) error {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == item.SourceCardID {
					if g.Battlefield.Cards[i].Counters == nil {
						g.Battlefield.Cards[i].Counters = map[string]int{}
					}
					g.Battlefield.Cards[i].Counters["effect-ran"]++
				}
			}
			return nil
		},
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func loyaltyOf(g *Game, id uuid.UUID) int {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return g.Battlefield.Cards[i].Counters["loyalty"]
		}
	}
	return -1
}

func counterOf(g *Game, id uuid.UUID, name string) int {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return g.Battlefield.Cards[i].Counters[name]
		}
	}
	return -1
}

// The #334 headline: a minus ability pays its loyalty, goes on the
// stack like any other CR 602 activation, and RUNS ITS EFFECT.
func TestLoyaltyAbilityPaysCostAndRunsEffect(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pw := pushLoyaltyWalker(g, me, 4, -3)

	if err := g.ActivateCatalogAbility(me.ID, pw, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if got := loyaltyOf(g, pw); got != 1 {
		t.Errorf("loyalty after -3 on a 4-loyalty walker: got %d, want 1", got)
	}
	// The ability is on the stack, not resolved yet (CR 602.2a).
	if counterOf(g, pw, "effect-ran") != 0 {
		t.Error("effect ran at announce; it belongs at resolution")
	}
	if len(g.StackMeta) != 1 {
		t.Fatalf("StackMeta has %d items, want 1", len(g.StackMeta))
	}
	_ = g.PassPriority()
	_ = g.PassPriority()
	if got := counterOf(g, pw, "effect-ran"); got != 1 {
		t.Errorf("effect-ran after resolution: got %d, want 1", got)
	}
}

// A plus ability adds counters rather than removing them.
func TestLoyaltyAbilityPlusAddsCounters(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pw := pushLoyaltyWalker(g, me, 4, 1)

	if err := g.ActivateCatalogAbility(me.ID, pw, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if got := loyaltyOf(g, pw); got != 5 {
		t.Errorf("loyalty after +1 on a 4-loyalty walker: got %d, want 5", got)
	}
}

// CR 606.6: you can't activate a loyalty ability whose cost removes
// more loyalty counters than the permanent has. Nothing is paid and
// nothing reaches the stack.
func TestLoyaltyAbilityCannotOverpayCR6063(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pw := pushLoyaltyWalker(g, me, 2, -3)

	if err := g.ActivateCatalogAbility(me.ID, pw, 0, ActivateAbilityParams{}); err != ErrInsufficientLoyalty {
		t.Errorf("-3 with 2 loyalty: got %v, want ErrInsufficientLoyalty", err)
	}
	if got := loyaltyOf(g, pw); got != 2 {
		t.Errorf("loyalty after a rejected activation: got %d, want 2", got)
	}
	if len(g.StackMeta) != 0 {
		t.Errorf("a rejected activation put %d items on the stack", len(g.StackMeta))
	}
	if g.LoyaltyActivatedThisTurn[pw] {
		t.Error("a rejected activation burned the once-per-turn window")
	}
}

// Paying a walker down to exactly 0 is legal (CR 606.6 forbids
// paying MORE than you have, not all of it); 704.5i then sweeps it.
func TestLoyaltyAbilityMayPayDownToZero(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pw := pushLoyaltyWalker(g, me, 3, -3)

	if err := g.ActivateCatalogAbility(me.ID, pw, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("-3 with exactly 3 loyalty: %v", err)
	}
	_ = g.PassPriority()
	_ = g.PassPriority()
	if g.Battlefield.Contains(pw) {
		t.Error("a 0-loyalty planeswalker survived the 704.5i SBA")
	}
	if !me.Graveyard.Contains(pw) {
		t.Error("the 0-loyalty planeswalker did not reach its owner's graveyard")
	}
}

// CR 606.3: one loyalty ability per planeswalker per turn, and the
// gate lives on the normal ActivateAbility path now, not only on the
// S13.1 sandbox action.
func TestLoyaltyAbilityOncePerTurn(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pw := pushLoyaltyWalker(g, me, 6, 1)

	if err := g.ActivateCatalogAbility(me.ID, pw, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	_ = g.PassPriority()
	_ = g.PassPriority()
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, ActivateAbilityParams{}); err != ErrLoyaltyAlreadyActivated {
		t.Errorf("second activation same turn: got %v, want ErrLoyaltyAlreadyActivated", err)
	}
	if got := loyaltyOf(g, pw); got != 7 {
		t.Errorf("loyalty after one +1: got %d, want 7", got)
	}
}

// A [0] ability changes no counters but still burns the turn's
// activation — which is why the cost is a *int and not an int.
func TestLoyaltyZeroCostStillGatesTheTurn(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pw := pushLoyaltyWalker(g, me, 3, 0)

	if err := g.ActivateCatalogAbility(me.ID, pw, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("[0] activation: %v", err)
	}
	if got := loyaltyOf(g, pw); got != 3 {
		t.Errorf("loyalty after [0]: got %d, want 3", got)
	}
	if !g.LoyaltyActivatedThisTurn[pw] {
		t.Error("[0] did not set the once-per-turn flag")
	}
}

// CR 606.3: loyalty abilities are sorcery-speed regardless of what
// the catalog entry says about SorcerySpeed — the loyalty component
// itself carries the restriction.
func TestLoyaltyAbilityIsSorcerySpeed(t *testing.T) {
	g := newActiveGame(t) // cursor sits on Upkeep
	me := g.Seats[0]
	pw := pushLoyaltyWalker(g, me, 4, 1)

	if err := g.ActivateCatalogAbility(me.ID, pw, 0, ActivateAbilityParams{}); err != ErrSorcerySpeedRequired {
		t.Errorf("upkeep activation: got %v, want ErrSorcerySpeedRequired", err)
	}
}

// CR 606.3 says a loyalty ability of a PERMANENT you control, and
// CR 606.1 defines the ability by the loyalty symbol in its cost —
// neither asks whether the permanent is a planeswalker. A permanent
// that carries one and is not a planeswalker is the printed case a
// type-setting effect makes (a Teferi under Song of the Dryads keeps
// his abilities and stops being a planeswalker), and this path used
// to refuse it with ErrNotAPlaneswalker on top of the controller
// check. #1157.
//
// The same activation is still gated on everything CR 606 DOES ask:
// sorcery speed, once per turn, enough counters for a −N — see the
// tests above and below, which run against a planeswalker because
// that is what nearly every loyalty ability is printed on.
func TestLoyaltyAbilityOnANonPlaneswalkerPermanent(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	c := NewCard("Sparkbound Relic", me.ID)
	c.TypeLine = "Artifact — Equipment"
	c.Controller = me.ID
	c.Counters = map[string]int{CounterLoyalty: 3}
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "+1: a loyalty ability on something that is not a planeswalker",
		Cost:  AbilityCost{Loyalty: loyaltyN(1)},
	}, {
		Label: "−5: more than it can pay",
		Cost:  AbilityCost{Loyalty: loyaltyN(-5)},
	}}
	g.Battlefield.PushTop(c)

	if err := g.ActivateCatalogAbility(me.ID, c.InstanceID, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("CR 606.3 is about a permanent, not a planeswalker: got %v", err)
	}
	if got := counterOf(g, c.InstanceID, CounterLoyalty); got != 4 {
		t.Errorf("loyalty after the +1: got %d, want 4", got)
	}
	// CR 606.3's once-per-turn clause says "that permanent" too. The
	// +1 resolves first: with it still on the stack the sorcery window
	// is shut (CR 307.1, #1352), and that is the refusal the second
	// activation would meet instead.
	resolveWholeStackForTest(t, g)
	if err := g.ActivateCatalogAbility(me.ID, c.InstanceID, 0, ActivateAbilityParams{}); err != ErrLoyaltyAlreadyActivated {
		t.Errorf("second activation the same turn: got %v, want ErrLoyaltyAlreadyActivated", err)
	}
}

// CR 606.5 is about the permanent's counters and nothing else, so it
// answers the same way off a planeswalker.
func TestLoyaltyMinusCostOnANonPlaneswalkerStillNeedsTheCounters(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	c := NewCard("Sparkbound Relic", me.ID)
	c.TypeLine = "Artifact — Equipment"
	c.Controller = me.ID
	c.Counters = map[string]int{CounterLoyalty: 3}
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "−5: more than it can pay",
		Cost:  AbilityCost{Loyalty: loyaltyN(-5)},
	}}
	g.Battlefield.PushTop(c)

	if err := g.ActivateCatalogAbility(me.ID, c.InstanceID, 0, ActivateAbilityParams{}); err != ErrInsufficientLoyalty {
		t.Errorf("−5 with three counters: got %v, want ErrInsufficientLoyalty", err)
	}
}

// CR 606.3: only the planeswalker's controller may activate it.
func TestLoyaltyAbilityRequiresControl(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, them := g.Seats[0], g.Seats[1]
	pw := pushLoyaltyWalker(g, them, 4, 1)

	if err := g.ActivateCatalogAbility(me.ID, pw, 0, ActivateAbilityParams{}); err != ErrCardCallerMismatch {
		t.Errorf("activating an opponent's walker: got %v, want ErrCardCallerMismatch", err)
	}
	if got := loyaltyOf(g, pw); got != 4 {
		t.Errorf("opponent's loyalty moved: got %d, want 4", got)
	}
}

// --- the S13.1 sandbox action, hardened --------------------------
//
// ActivateLoyalty stays as the manual path for the ~thousand
// planeswalkers with no catalog entry, but it used to take any
// battlefield card, from any player, for any delta.

func TestSandboxActivateLoyaltyRejectsNonPlaneswalker(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	bear := pushIntrinsicPermanent(g, me, "Bear", "Creature — Bear", nil, nil)

	if err := g.ActivateLoyalty(me.ID, bear, "+1", 1); err != ErrNotAPlaneswalker {
		t.Errorf("loyalty on a Bear: got %v, want ErrNotAPlaneswalker", err)
	}
	if got := counterOf(g, bear, "loyalty"); got != 0 {
		t.Errorf("the Bear grew %d loyalty counters", got)
	}
}

func TestSandboxActivateLoyaltyRequiresControl(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, them := g.Seats[0], g.Seats[1]
	pw := pushLoyaltyWalker(g, them, 4, 1)

	if err := g.ActivateLoyalty(me.ID, pw, "+1", 1); err != ErrCardCallerMismatch {
		t.Errorf("sandbox activation of an opponent's walker: got %v, want ErrCardCallerMismatch", err)
	}
}

func TestSandboxActivateLoyaltyEnforcesCR6063(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pw := pushLoyaltyWalker(g, me, 2, 0)

	if err := g.ActivateLoyalty(me.ID, pw, "-5", -5); err != ErrInsufficientLoyalty {
		t.Errorf("sandbox -5 with 2 loyalty: got %v, want ErrInsufficientLoyalty", err)
	}
	if got := loyaltyOf(g, pw); got != 2 {
		t.Errorf("loyalty after a rejected sandbox activation: got %d, want 2", got)
	}
	if g.LoyaltyActivatedThisTurn[pw] {
		t.Error("a rejected sandbox activation burned the once-per-turn window")
	}
}
