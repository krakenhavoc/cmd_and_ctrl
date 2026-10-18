package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wayfarer's Bauble: "{2}, {T}, Sacrifice this artifact: Search your
// library for a basic land card, put that card onto the battlefield
// tapped, then shuffle." The shape test in staples2_test.go pins the cost
// components; these pin what the ability DOES when it resolves.

func baubleTable(t *testing.T) (g *game.Game, me *game.Player, bauble, basic, nonbasic uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		basic = pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
		nonbasic = pushLibraryCardForTest(me, game.Card{Name: "Breeding Pool", TypeLine: "Land — Forest Island"})
	})
	bauble = pushCatalogPermanent(g, me.ID, "Wayfarer's Bauble", "Artifact", wayfarersBaubleOracle, false)
	return g, me, bauble, basic, nonbasic
}

func TestWayfarersBaubleFetchesATappedBasicAndIsSacrificed(t *testing.T) {
	g, me, bauble, basic, nonbasic := baubleTable(t)
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bauble, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(bauble) {
		t.Error("the Bauble is still on the battlefield: sacrificing it is a cost")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	for _, id := range c.SearchCards {
		if id == nonbasic {
			t.Error("the search offered a nonbasic land: the Bauble finds a BASIC land card")
		}
	}
	answerSearchByID(t, g, me.ID, basic)
	got, ok := battlefieldCard(g, basic)
	if !ok {
		t.Fatal("the basic land is not on the battlefield")
	}
	if !got.Tapped {
		t.Error("the fetched land must enter tapped")
	}
}

func TestWayfarersBaubleNeedsTheTapAndTheMana(t *testing.T) {
	g, me, bauble, _, _ := baubleTable(t)
	if err := g.ActivateCatalogAbility(me.ID, bauble, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the Bauble activated with no mana to pay {2}")
	}
	floatMana(t, g, me, "{C}{C}")
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bauble) })
	if err := g.ActivateCatalogAbility(me.ID, bauble, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a tapped Bauble activated: {T} is part of the cost")
	}
}
