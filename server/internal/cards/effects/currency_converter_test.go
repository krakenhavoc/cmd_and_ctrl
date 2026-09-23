package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// currency_converter_test.go — #1218: the card proof that
// b27ExiledWith plus game.ChooseCardsPrompt are enough to ship this
// card, with no new engine seam. See the doc comment on
// currency_converter.go.

// chooseCardsPromptFrom returns the choose-cards prompt raised by
// `source`, or nil. Distinguished from an ordinary discard prompt
// (also a PendingChoiceChooseCards) by Source rather than by
// Chooser/FromPlayer, which a hand-pick discard shares by shape.
func chooseCardsPromptFrom(g *game.Game, source uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Source == source {
			return c
		}
	}
	return nil
}

func TestCurrencyConverterExilesADiscardedCardWhenYouSayYes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedPermanentWithOracle(g, me.ID, "Currency Converter", "Artifact", currencyConverterOracle)
	emptyHandToLibrary(g, me)
	pitched := handCardFull(me, "Big Wurm", "Creature — Wurm", "{5}{G}", "", nil)

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(pitched) {
		t.Fatal("the discarded card should be exiled from the graveyard")
	}
	if me.Graveyard.Contains(pitched) {
		t.Error("the discarded card should not stay in the graveyard")
	}
}

func TestCurrencyConverterDeclinesToExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedPermanentWithOracle(g, me.ID, "Currency Converter", "Artifact", currencyConverterOracle)
	emptyHandToLibrary(g, me)
	pitched := handCardFull(me, "Big Wurm", "Creature — Wurm", "{5}{G}", "", nil)

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	if g.Exile.Contains(pitched) {
		t.Error("declining the trigger should leave the card in the graveyard")
	}
	if !me.Graveyard.Contains(pitched) {
		t.Error("the discarded card should stay in the graveyard")
	}
}

func TestCurrencyConverterThirdAbilityMakesATreasureForALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cc := seedPermanentWithOracle(g, me.ID, "Currency Converter", "Artifact", currencyConverterOracle)
	emptyHandToLibrary(g, me)
	pitched := handCardFull(me, "Stashed Land", "Land", "", "", nil)

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(pitched) {
		t.Fatalf("setup: the discarded land should be exiled")
	}

	if err := g.ActivateCatalogAbility(me.ID, cc, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the third ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := chooseCardsPromptFrom(g, cc)
	if c == nil {
		t.Fatalf("no choose-cards prompt from Currency Converter")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{pitched}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(pitched) {
		t.Error("the chosen card should be in its owner's graveyard")
	}
	if findBattlefieldByName(g, "Treasure") == uuid.Nil {
		t.Error("a land card should make a Treasure token")
	}
}

func TestCurrencyConverterThirdAbilityMakesARogueForANonland(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cc := seedPermanentWithOracle(g, me.ID, "Currency Converter", "Artifact", currencyConverterOracle)
	emptyHandToLibrary(g, me)
	pitched := handCardFull(me, "Big Wurm", "Creature — Wurm", "{5}{G}", "", nil)

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(pitched) {
		t.Fatalf("setup: the discarded creature should be exiled")
	}

	if err := g.ActivateCatalogAbility(me.ID, cc, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the third ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := chooseCardsPromptFrom(g, cc)
	if c == nil {
		t.Fatalf("no choose-cards prompt from Currency Converter")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{pitched}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(pitched) {
		t.Error("the chosen card should be in its owner's graveyard")
	}
	rogue := findBattlefieldByName(g, "Rogue")
	if rogue == uuid.Nil {
		t.Fatal("a nonland card should make a Rogue token")
	}
	if c, ok := g.LookupCardForEffect(rogue); !ok || c.Power != 2 || c.Toughness != 2 {
		t.Errorf("the Rogue token should be 2/2, got %+v", c)
	}
}

func TestCurrencyConverterThirdAbilityNoOpsWithNothingExiled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cc := seedPermanentWithOracle(g, me.ID, "Currency Converter", "Artifact", currencyConverterOracle)

	if err := g.ActivateCatalogAbility(me.ID, cc, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the third ability with nothing exiled: %v", err)
	}
	passPriorityAroundTable(t, g)

	if c := chooseCardsPromptFrom(g, cc); c != nil {
		t.Error("no choice should be offered with nothing exiled")
	}
	if got, ok := g.LookupCardForEffect(cc); !ok || !got.Tapped {
		t.Error("the ability still costs the tap even with nothing to choose")
	}
}
