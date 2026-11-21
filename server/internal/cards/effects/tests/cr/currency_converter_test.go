package testscr

import (
	"testing"

	"github.com/google/uuid"

	cardsco "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/co"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestCurrencyConverterExilesADiscardedCardWhenYouSayYes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedPermanentWithOracle(g, me.ID, "Currency Converter", "Artifact", cardsco.CurrencyConverterOracle)
	effectstest.EmptyHandToLibrary(g, me)
	pitched := effectstest.HandCardFull(me, "Big Wurm", "Creature — Wurm", "{5}{G}", "", nil)

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(pitched) {
		t.Fatal("the discarded card should be exiled from the graveyard")
	}
	if me.Graveyard.Contains(pitched) {
		t.Error("the discarded card should not stay in the graveyard")
	}
}

func TestCurrencyConverterDeclinesToExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedPermanentWithOracle(g, me.ID, "Currency Converter", "Artifact", cardsco.CurrencyConverterOracle)
	effectstest.EmptyHandToLibrary(g, me)
	pitched := effectstest.HandCardFull(me, "Big Wurm", "Creature — Wurm", "{5}{G}", "", nil)

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Exile.Contains(pitched) {
		t.Error("declining the trigger should leave the card in the graveyard")
	}
	if !me.Graveyard.Contains(pitched) {
		t.Error("the discarded card should stay in the graveyard")
	}
}

func TestCurrencyConverterThirdAbilityMakesATreasureForALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cc := effectstest.SeedPermanentWithOracle(g, me.ID, "Currency Converter", "Artifact", cardsco.CurrencyConverterOracle)
	effectstest.EmptyHandToLibrary(g, me)
	pitched := effectstest.HandCardFull(me, "Stashed Land", "Land", "", "", nil)

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(pitched) {
		t.Fatalf("setup: the discarded land should be exiled")
	}

	if err := g.ActivateCatalogAbility(me.ID, cc, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the third ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsPromptFrom(g, cc)
	if c == nil {
		t.Fatalf("no choose-cards prompt from Currency Converter")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{pitched}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(pitched) {
		t.Error("the chosen card should be in its owner's graveyard")
	}
	if effectstest.FindBattlefieldByName(g, "Treasure") == uuid.Nil {
		t.Error("a land card should make a Treasure token")
	}
}

func TestCurrencyConverterThirdAbilityMakesARogueForANonland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cc := effectstest.SeedPermanentWithOracle(g, me.ID, "Currency Converter", "Artifact", cardsco.CurrencyConverterOracle)
	effectstest.EmptyHandToLibrary(g, me)
	pitched := effectstest.HandCardFull(me, "Big Wurm", "Creature — Wurm", "{5}{G}", "", nil)

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(pitched) {
		t.Fatalf("setup: the discarded creature should be exiled")
	}

	if err := g.ActivateCatalogAbility(me.ID, cc, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the third ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsPromptFrom(g, cc)
	if c == nil {
		t.Fatalf("no choose-cards prompt from Currency Converter")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{pitched}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(pitched) {
		t.Error("the chosen card should be in its owner's graveyard")
	}
	rogue := effectstest.FindBattlefieldByName(g, "Rogue")
	if rogue == uuid.Nil {
		t.Fatal("a nonland card should make a Rogue token")
	}
	if c, ok := g.LookupCardForEffect(rogue); !ok || c.Power != 2 || c.Toughness != 2 {
		t.Errorf("the Rogue token should be 2/2, got %+v", c)
	}
}

func TestCurrencyConverterThirdAbilityNoOpsWithNothingExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cc := effectstest.SeedPermanentWithOracle(g, me.ID, "Currency Converter", "Artifact", cardsco.CurrencyConverterOracle)

	if err := g.ActivateCatalogAbility(me.ID, cc, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the third ability with nothing exiled: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if c := effectstest.ChooseCardsPromptFrom(g, cc); c != nil {
		t.Error("no choice should be offered with nothing exiled")
	}
	if got, ok := g.LookupCardForEffect(cc); !ok || !got.Tapped {
		t.Error("the ability still costs the tap even with nothing to choose")
	}
}
