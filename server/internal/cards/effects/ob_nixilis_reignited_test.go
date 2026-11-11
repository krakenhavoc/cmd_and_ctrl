package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const obNixilisReignitedOracle = "65481ae1-54b2-44d2-91e1-20cd486386da"

func playerRefs(ids ...uuid.UUID) []game.TargetRef {
	var out []game.TargetRef
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetPlayer, ID: id})
	}
	return out
}

func TestObNixilisReignitedPlusDrawsAndLosesLife(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pw := pushWalkerForTest(g, me.ID, "Ob Nixilis Reignited", obNixilisReignitedOracle, 5)
	advanceToMainOf(t, g, seat)
	hand, life := handSize(me), lifeOf(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != hand+1 {
		t.Errorf("hand = %d, want %d", got, hand+1)
	}
	if got := lifeOf(g, me.ID); got != life-1 {
		t.Errorf("life = %d, want %d", got, life-1)
	}
	if got := loyaltyOf(g, pw); got != 6 {
		t.Errorf("loyalty = %d, want 6", got)
	}
}

func TestObNixilisReignitedMinusThreeDestroysACreature(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	pw := pushWalkerForTest(g, me.ID, "Ob Nixilis Reignited", obNixilisReignitedOracle, 5)
	victim := pushCreatureToBattlefieldForTest(g, opp.ID, "Victim")
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Targets: cardRefs(pw), Strict: true}); err == nil {
		t.Fatal("−3 targeting a planeswalker was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Targets: cardRefs(victim), Strict: true}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == victim {
			t.Fatal("the creature survived")
		}
	}
	if got := loyaltyOf(g, pw); got != 2 {
		t.Errorf("loyalty = %d, want 2", got)
	}
}

func TestObNixilisReignitedEmblemPunishesTheTargetOnEveryDraw(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	other := g.Seats[(seat+2)%len(g.Seats)]
	pw := pushWalkerForTest(g, me.ID, "Ob Nixilis Reignited", obNixilisReignitedOracle, 8)
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{Targets: playerRefs(me.ID), Strict: true}); err == nil {
		t.Fatal("−8 targeting its own controller was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{Targets: playerRefs(opp.ID), Strict: true}); err != nil {
		t.Fatalf("−8: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(emblemsOf(g, opp.ID)) != 1 || len(emblemsOf(g, me.ID)) != 0 {
		t.Fatalf("emblems: opponent %d, me %d; want 1 and 0", len(emblemsOf(g, opp.ID)), len(emblemsOf(g, me.ID)))
	}
	before := lifeOf(g, opp.ID)
	for _, drawer := range []*game.Player{me, other, opp} {
		if err := g.DrawCard(drawer.ID); err != nil {
			t.Fatalf("DrawCard: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if got := lifeOf(g, opp.ID); got != before-6 {
		t.Errorf("opponent life = %d, want %d (2 per draw by any player)", got, before-6)
	}
}
