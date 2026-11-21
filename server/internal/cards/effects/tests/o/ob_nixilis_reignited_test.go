package testso

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const obNixilisReignitedOracle = "65481ae1-54b2-44d2-91e1-20cd486386da"

func TestObNixilisReignitedPlusDrawsAndLosesLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pw := effectstest.PushWalkerForTest(g, me.ID, "Ob Nixilis Reignited", obNixilisReignitedOracle, 5)
	effectstest.AdvanceToMainOf(t, g, seat)
	hand, life := effectstest.HandSize(me), effectstest.LifeOf(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSize(me); got != hand+1 {
		t.Errorf("hand = %d, want %d", got, hand+1)
	}
	if got := effectstest.LifeOf(g, me.ID); got != life-1 {
		t.Errorf("life = %d, want %d", got, life-1)
	}
	if got := effectstest.LoyaltyOf(g, pw); got != 6 {
		t.Errorf("loyalty = %d, want 6", got)
	}
}

func TestObNixilisReignitedMinusThreeDestroysACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	pw := effectstest.PushWalkerForTest(g, me.ID, "Ob Nixilis Reignited", obNixilisReignitedOracle, 5)
	victim := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Victim")
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(pw), Strict: true}); err == nil {
		t.Fatal("−3 targeting a planeswalker was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim), Strict: true}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == victim {
			t.Fatal("the creature survived")
		}
	}
	if got := effectstest.LoyaltyOf(g, pw); got != 2 {
		t.Errorf("loyalty = %d, want 2", got)
	}
}

func TestObNixilisReignitedEmblemPunishesTheTargetOnEveryDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	other := g.Seats[(seat+2)%len(g.Seats)]
	pw := effectstest.PushWalkerForTest(g, me.ID, "Ob Nixilis Reignited", obNixilisReignitedOracle, 8)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(me.ID), Strict: true}); err == nil {
		t.Fatal("−8 targeting its own controller was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID), Strict: true}); err != nil {
		t.Fatalf("−8: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.EmblemsOf(g, opp.ID)) != 1 || len(effectstest.EmblemsOf(g, me.ID)) != 0 {
		t.Fatalf("emblems: opponent %d, me %d; want 1 and 0", len(effectstest.EmblemsOf(g, opp.ID)), len(effectstest.EmblemsOf(g, me.ID)))
	}
	before := effectstest.LifeOf(g, opp.ID)
	for _, drawer := range []*game.Player{me, other, opp} {
		if err := g.DrawCard(drawer.ID); err != nil {
			t.Fatalf("DrawCard: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.LifeOf(g, opp.ID); got != before-6 {
		t.Errorf("opponent life = %d, want %d (2 per draw by any player)", got, before-6)
	}
}
