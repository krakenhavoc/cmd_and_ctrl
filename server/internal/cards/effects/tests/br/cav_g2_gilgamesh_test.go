package testsbr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// Gilgamesh's reflexive sentence: once Equipment entered, a second trigger
// goes on the stack, and on resolution the controller picks one of them to
// attach to a Samurai (Gilgamesh is one himself).
func TestGilgameshOffersToAttachOneEquipmentToASamurai(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sword := effectstest.PlTop(me, "Sword", "Artifact — Equipment", "{3}")
	axe := effectstest.PlTop(me, "Axe", "Artifact — Equipment", "{2}")
	gil := effectstest.CastCatalogSpell(t, g, "Gilgamesh, Master-at-Arms", "Legendary Creature — Human Samurai", effectstest.PlGilgameshOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AnswerChooseCards(t, g, me.ID, sword, axe)
	if effectstest.TriggerOnStack(g, gil) == nil {
		t.Fatal("the reflexive attach trigger should be on the stack, a response window")
	}
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || !effectstest.HasID(pick.ChooseCards, sword) || !effectstest.HasID(pick.ChooseCards, axe) || pick.ChooseMin != 0 {
		t.Fatalf("want an optional pick over the two entered Equipment, got %+v", pick)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, axe)
	host := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if host == nil || len(host.ChooseCards) != 1 || host.ChooseCards[0] != gil {
		t.Fatalf("want Gilgamesh as the only Samurai, got %+v", host)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, gil)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sword := effectstest.PlTop(me, "Sword", "Artifact — Equipment", "{3}")
	gil := effectstest.CastCatalogSpell(t, g, "Gilgamesh, Master-at-Arms", "Legendary Creature — Human Samurai", effectstest.PlGilgameshOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, sword)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	swordCard, _ := g.LookupCardForEffect(sword)
	if swordCard.IsAttachedTo(gil) || g.Battlefield.Contains(sword) == false {
		t.Errorf("declined: want the Sword on the battlefield, unattached")
	}
}

// Putting no Equipment onto the battlefield creates no reflexive trigger.
func TestGilgameshWithoutEquipmentHasNothingToAttach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PlTop(me, "Bear", "Creature — Bear", "{1}{G}")
	effectstest.CastCatalogSpell(t, g, "Gilgamesh, Master-at-Arms", "Legendary Creature — Human Samurai", effectstest.PlGilgameshOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.ChooseCardsChoiceFor(g, me.ID); pick != nil {
		t.Fatalf("no Equipment among the six: no pick expected, got %+v", pick)
	}
}
