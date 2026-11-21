package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kiora_of_salt_and_sand_test.go — Kiora of Salt and Sand's trigger half
// (#2797). The granted −8 is proved with the class grant in
// planeswalker_statics_2797_test.go; this file is "Whenever you attack, if
// you've activated a loyalty ability this turn, untap target attacking
// creature. It can't be blocked this turn."

func TestKioraIsRegisteredAndWhole(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.KioraOracle2797)
	if !ok {
		t.Fatal("Kiora of Salt and Sand is not registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %v, want Full and none", spec.Completeness, spec.Caveats)
	}
	if len(spec.Grants) != 1 || len(spec.Static) != 1 || len(spec.Triggered) != 1 {
		t.Errorf("grants=%d statics=%d triggers=%d, want 1/1/1", len(spec.Grants), len(spec.Static), len(spec.Triggered))
	}
}

// One trigger for the whole attack, not one per attacker; the chosen attacker
// untaps and can't be blocked, the other does neither.
func TestKioraUntapsAndUnblocksOneAttackerAfterALoyaltyActivation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Kiora of Salt and Sand", "Legendary Creature — Merfolk Noble", effectstest.KioraOracle2797, false)
	walker := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 4)
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)

	effectstest.B16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatalf("first attack: %v", err)
	}
	if err := g.DeclareAttacker(b, opp.ID); err != nil {
		t.Fatalf("second attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerPickTarget(t, g, a)
	effectstest.PassPriorityAroundTable(t, g)

	if tapped, _ := effectstest.BattlefieldCardTapped(g, a); tapped {
		t.Error("the chosen attacker is still tapped")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, b); !tapped {
		t.Error("the other attacker untapped — the trigger names one creature")
	}
	if effectstest.RestrictionsOf(t, g, a)&game.CantBeBlocked == 0 {
		t.Error("the chosen attacker can still be blocked")
	}
	if effectstest.RestrictionsOf(t, g, b)&game.CantBeBlocked != 0 {
		t.Error("the other attacker can't be blocked either")
	}
}

// The intervening "if" (CR 603.4): with no loyalty ability activated this turn
// the trigger never goes on the stack, so there is no target prompt at all.
func TestKioraDoesNothingWithoutAnActivatedLoyaltyAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Kiora of Salt and Sand", "Legendary Creature — Merfolk Noble", effectstest.KioraOracle2797, false)
	effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 4)
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("Kiora asked for a target though no loyalty ability was activated this turn")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RestrictionsOf(t, g, a)&game.CantBeBlocked != 0 {
		t.Error("the attacker can't be blocked without a loyalty activation")
	}
}

// "You've activated" — an opponent's loyalty activation does not count, and a
// granted row does.
func TestKioraCountsYourActivationsOfGrantedRowsAndNotAnOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Kiora of Salt and Sand", "Legendary Creature — Merfolk Noble", effectstest.KioraOracle2797, false)

	if effects.YouActivatedALoyaltyAbilityThisTurn(g, me.ID) {
		t.Fatal("setup: I have activated nothing")
	}
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventActivateAbility, Actor: opp.ID, Loyalty: true, CardID: uuid.New()})
	})
	if effects.YouActivatedALoyaltyAbilityThisTurn(g, me.ID) {
		t.Error("an opponent's loyalty activation counted as mine")
	}

	walker := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, "Planeswalker — Fix", 8)
	idx, ref := effectstest.GrantedLoyaltyRow(t, g, walker, "−8")
	effectstest.B16Activate(t, g, me.ID, walker, idx, game.ActivateAbilityParams{Ref: ref})
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.YouActivatedALoyaltyAbilityThisTurn(g, me.ID) {
		t.Error("my activation of a granted loyalty row did not count")
	}
}
