package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kiora_of_salt_and_sand_test.go — Kiora of Salt and Sand's trigger half
// (#2797). The granted −8 is proved with the class grant in
// planeswalker_statics_2797_test.go; this file is "Whenever you attack, if
// you've activated a loyalty ability this turn, untap target attacking
// creature. It can't be blocked this turn."

func TestKioraIsRegisteredAndWhole(t *testing.T) {
	spec, ok := Lookup(kioraOracle2797)
	if !ok {
		t.Fatal("Kiora of Salt and Sand is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %v, want Full and none", spec.Completeness, spec.Caveats)
	}
	if len(spec.Grants) != 1 || len(spec.Static) != 1 || len(spec.Triggered) != 1 {
		t.Errorf("grants=%d statics=%d triggers=%d, want 1/1/1", len(spec.Grants), len(spec.Static), len(spec.Triggered))
	}
}

// One trigger for the whole attack, not one per attacker; the chosen attacker
// untaps and can't be blocked, the other does neither.
func TestKioraUntapsAndUnblocksOneAttackerAfterALoyaltyActivation(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Kiora of Salt and Sand", "Legendary Creature — Merfolk Noble", kioraOracle2797, false)
	walker := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 4)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)

	b16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatalf("first attack: %v", err)
	}
	if err := g.DeclareAttacker(b, opp.ID); err != nil {
		t.Fatalf("second attack: %v", err)
	}
	lockInAttacks(t, g)
	answerPickTarget(t, g, a)
	passPriorityAroundTable(t, g)

	if tapped, _ := battlefieldCardTapped(g, a); tapped {
		t.Error("the chosen attacker is still tapped")
	}
	if tapped, _ := battlefieldCardTapped(g, b); !tapped {
		t.Error("the other attacker untapped — the trigger names one creature")
	}
	if restrictionsOf(t, g, a)&game.CantBeBlocked == 0 {
		t.Error("the chosen attacker can still be blocked")
	}
	if restrictionsOf(t, g, b)&game.CantBeBlocked != 0 {
		t.Error("the other attacker can't be blocked either")
	}
}

// The intervening "if" (CR 603.4): with no loyalty ability activated this turn
// the trigger never goes on the stack, so there is no target prompt at all.
func TestKioraDoesNothingWithoutAnActivatedLoyaltyAbility(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Kiora of Salt and Sand", "Legendary Creature — Merfolk Noble", kioraOracle2797, false)
	fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 4)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	lockInAttacks(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("Kiora asked for a target though no loyalty ability was activated this turn")
	}
	passPriorityAroundTable(t, g)
	if restrictionsOf(t, g, a)&game.CantBeBlocked != 0 {
		t.Error("the attacker can't be blocked without a loyalty activation")
	}
}

// "You've activated" — an opponent's loyalty activation does not count, and a
// granted row does.
func TestKioraCountsYourActivationsOfGrantedRowsAndNotAnOpponents(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Kiora of Salt and Sand", "Legendary Creature — Merfolk Noble", kioraOracle2797, false)

	if youActivatedALoyaltyAbilityThisTurn(g, me.ID) {
		t.Fatal("setup: I have activated nothing")
	}
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventActivateAbility, Actor: opp.ID, Loyalty: true, CardID: uuid.New()})
	})
	if youActivatedALoyaltyAbilityThisTurn(g, me.ID) {
		t.Error("an opponent's loyalty activation counted as mine")
	}

	walker := fixtureWalker(g, me.ID, pw2797WalkerOracle, "Planeswalker — Fix", 8)
	idx, ref := grantedLoyaltyRow(t, g, walker, "−8")
	b16Activate(t, g, me.ID, walker, idx, game.ActivateAbilityParams{Ref: ref})
	passPriorityAroundTable(t, g)
	if !youActivatedALoyaltyAbilityThisTurn(g, me.ID) {
		t.Error("my activation of a granted loyalty row did not count")
	}
}
