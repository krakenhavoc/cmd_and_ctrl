package effects

import (
	"testing"
)

// Gilgamesh's reflexive sentence: once Equipment entered, a second trigger
// goes on the stack, and on resolution the controller picks one of them to
// attach to a Samurai (Gilgamesh is one himself).
func TestGilgameshOffersToAttachOneEquipmentToASamurai(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sword := plTop(me, "Sword", "Artifact — Equipment", "{3}")
	axe := plTop(me, "Axe", "Artifact — Equipment", "{2}")
	gil := castCatalogSpell(t, g, "Gilgamesh, Master-at-Arms", "Legendary Creature — Human Samurai", plGilgameshOracle, nil)
	passPriorityAroundTable(t, g)

	answerChooseCards(t, g, me.ID, sword, axe)
	if triggerOnStack(g, gil) == nil {
		t.Fatal("the reflexive attach trigger should be on the stack, a response window")
	}
	passPriorityAroundTable(t, g)

	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil || !hasID(pick.ChooseCards, sword) || !hasID(pick.ChooseCards, axe) || pick.ChooseMin != 0 {
		t.Fatalf("want an optional pick over the two entered Equipment, got %+v", pick)
	}
	answerChooseCards(t, g, me.ID, axe)
	host := chooseCardsChoiceFor(g, me.ID)
	if host == nil || len(host.ChooseCards) != 1 || host.ChooseCards[0] != gil {
		t.Fatalf("want Gilgamesh as the only Samurai, got %+v", host)
	}
	answerChooseCards(t, g, me.ID, gil)

	axeCard, _ := g.LookupCardForEffect(axe)
	swordCard, _ := g.LookupCardForEffect(sword)
	if !axeCard.IsAttachedTo(gil) {
		t.Error("the chosen Equipment should be attached to Gilgamesh")
	}
	if swordCard.IsAttachedTo(gil) {
		t.Error("only one Equipment may be attached")
	}
}

// Declining the optional pick leaves every Equipment unattached.
func TestGilgameshAttachCanBeDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sword := plTop(me, "Sword", "Artifact — Equipment", "{3}")
	gil := castCatalogSpell(t, g, "Gilgamesh, Master-at-Arms", "Legendary Creature — Human Samurai", plGilgameshOracle, nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, sword)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	swordCard, _ := g.LookupCardForEffect(sword)
	if swordCard.IsAttachedTo(gil) || g.Battlefield.Contains(sword) == false {
		t.Errorf("declined: want the Sword on the battlefield, unattached")
	}
}

// Putting no Equipment onto the battlefield creates no reflexive trigger.
func TestGilgameshWithoutEquipmentHasNothingToAttach(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	plTop(me, "Bear", "Creature — Bear", "{1}{G}")
	castCatalogSpell(t, g, "Gilgamesh, Master-at-Arms", "Legendary Creature — Human Samurai", plGilgameshOracle, nil)
	passPriorityAroundTable(t, g)
	if pick := chooseCardsChoiceFor(g, me.ID); pick != nil {
		t.Fatalf("no Equipment among the six: no pick expected, got %+v", pick)
	}
}
