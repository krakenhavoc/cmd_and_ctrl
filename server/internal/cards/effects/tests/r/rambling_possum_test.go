package testsr

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// Rambling Possum is the card that reads "creatures that saddled it
// this turn" (CR 702.171c): the prompt offers exactly the saddlers still
// on the battlefield, "any number" includes none, and the pump does not
// wait for the answer.
func TestRamblingPossumOffersTheSaddlersAndReturnsTheChosenOnes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := effectstest.PushMountForTest(g, me.ID, "Rambling Possum", effectstest.RamblingPossumOracle, 2)
	keep := effectstest.PushCrewerForTest(g, me.ID, "Keeper", 1)
	bounce := effectstest.PushCrewerForTest(g, me.ID, "Bounce Me", 1)
	bystander := effectstest.PushCrewerForTest(g, me.ID, "Bystander", 1)
	effectstest.ToMain(t, g)
	if err := effectstest.SaddleAbility(t, g, me.ID, mount, keep, bounce); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, mount)
	effectstest.PassPriorityAroundTable(t, g)

	// The pump is already on.
	if got := effectstest.EffectivePower(t, g, mount); got != 3 {
		t.Errorf("power = %d, want 3 (2 + 1)", got)
	}
	choice := effectstest.LatestChooseCards(g, me.ID)
	if choice == nil {
		t.Fatal("no prompt for the creatures that saddled it")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range choice.ChooseCards {
		offered[id] = true
	}
	if !offered[keep] || !offered[bounce] || len(offered) != 2 {
		t.Errorf("offered %v, want exactly the two saddlers (not the bystander %s)", choice.ChooseCards, bystander)
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{bounce}); err != nil {
		t.Fatalf("answer: %v", err)
	}
	if _, ok := effectstest.BattlefieldCardByID(g, bounce); ok {
		t.Error("the chosen saddler is still on the battlefield")
	}
	if !me.Hand.Contains(bounce) {
		t.Error("the chosen saddler did not return to its owner's hand")
	}
	for _, id := range []uuid.UUID{keep, bystander, mount} {
		if _, ok := effectstest.BattlefieldCardByID(g, id); !ok {
			t.Errorf("%s left the battlefield; only the chosen saddler should", id)
		}
	}
}

// Choosing none is a legal answer.
func TestRamblingPossumMayReturnNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := effectstest.PushMountForTest(g, me.ID, "Rambling Possum", effectstest.RamblingPossumOracle, 2)
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Saddler", 1)
	effectstest.ToMain(t, g)
	if err := effectstest.SaddleAbility(t, g, me.ID, mount, saddler); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, mount)
	effectstest.PassPriorityAroundTable(t, g)
	choice := effectstest.LatestChooseCards(g, me.ID)
	if choice == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, nil); err != nil {
		t.Fatalf("answering with nothing: %v", err)
	}
	if _, ok := effectstest.BattlefieldCardByID(g, saddler); !ok {
		t.Error("the saddler left the battlefield though none was chosen")
	}
}
