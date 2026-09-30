package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const lootOracle = "f6476884-f73f-465f-8e7f-ea312a5b306d"

// TestLootGrantsAnAdditionalLandDrop — a plain AdditionalLandPlays
// read: with Loot on the battlefield the controller has two land
// drops instead of one.
func TestLootGrantsAnAdditionalLandDrop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := g.LandDropsRemainingFor(me.ID)
	pushCatalogPermanent(g, me.ID, "Loot, Exuberant Explorer", "Legendary Creature — Beast Noble", lootOracle, false)
	if got := g.LandDropsRemainingFor(me.ID); got != before+1 {
		t.Errorf("land drops with Loot in play: %d, want %d", got, before+1)
	}
}

// TestLootTapAbilityPutsAnAffordableCreatureOntoTheBattlefield — the
// top-six look offers only the creature within the land-count budget;
// the rest go to the bottom.
func TestLootTapAbilityPutsAnAffordableCreatureOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	loot := pushCatalogPermanent(g, me.ID, "Loot, Exuberant Explorer", "Legendary Creature — Beast Noble", lootOracle, false)
	// Two lands controlled.
	pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)

	g.WithWriteLock(func() { me.Library.Cards = nil })
	libSizeBefore := me.Library.Size()
	cheap := pushLibraryCardForTest(me, game.Card{Name: "Cheap Beast", TypeLine: "Creature — Beast", Power: 2, Toughness: 2, ManaCost: "{1}{G}"})
	expensive := pushLibraryCardForTest(me, game.Card{Name: "Expensive Beast", TypeLine: "Creature — Beast", Power: 9, Toughness: 9, ManaCost: "{7}{G}{G}"})
	for i := 0; i < 4; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	}

	floatMana(t, g, me, "{C}{C}{C}{C}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, loot, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no choose_cards prompt")
	}
	for _, id := range pick.ChooseCards {
		if id == expensive {
			t.Fatalf("the mana-value-9 creature should not be offered against two lands")
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{cheap}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(cheap) {
		t.Error("the affordable creature should be on the battlefield")
	}
	if me.Library.Contains(cheap) {
		t.Error("the chosen card should have left the library")
	}
	if got := me.Library.Size(); got != libSizeBefore+5 {
		t.Errorf("library size after look-6/take-1: %d, want %d (one taken, five bottomed)", got, libSizeBefore+5)
	}
}
