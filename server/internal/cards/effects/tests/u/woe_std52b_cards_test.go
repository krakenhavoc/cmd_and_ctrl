package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_std52b_cards_test.go — the second half of Standard batch 52's tests
// (#4471, #4452).

// --- Discerning Financier --------------------------------------------------

func TestStd52FinancierMakesATreasureOnlyWhenAnOpponentHasMoreLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Discerning Financier", "Creature — Human Noble", std52FinancierOracle, false)
	emit := func() {
		g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID}) })
		effectstest.PassPriorityAroundTable(t, g)
	}
	emit()
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Treasure"); n != 0 {
		t.Fatalf("%d Treasures with level lands, want 0", n)
	}
	effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	emit()
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasures with an opponent ahead on lands, want 1", n)
	}
}

func TestStd52FinancierGivesAChosenPlayerYourTreasureAndYouDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	financier := effectstest.PushCatalogPermanent(g, me.ID, "Discerning Financier", "Creature — Human Noble", std52FinancierOracle, false)
	effectstest.HobB10Treasures(g, me.ID, 1)
	treasure := effectstest.FindBattlefieldByName(g, "Treasure")
	std52ToMain(t, g)
	hand := effectstest.HandSize(me)
	effectstest.B06AddMana(me, "C", "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, financier, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(treasure), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, treasure); c.Controller != opp.ID {
		t.Errorf("the Treasure is controlled by %v, want the chosen player %v", c.Controller, opp.ID)
	}
	if got := effectstest.HandSize(me); got != hand+1 {
		t.Errorf("hand %d, want %d: you draw a card", got, hand+1)
	}
}

func TestStd52FinancierCannotTargetANonTreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	financier := effectstest.PushCatalogPermanent(g, me.ID, "Discerning Financier", "Creature — Human Noble", std52FinancierOracle, false)
	rock := effectstest.B12Permanent(g, me.ID, "Mind Stone", "Artifact")
	std52ToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, financier, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(rock), Strict: true}); err == nil {
		t.Error("a non-Treasure artifact was accepted as the target")
	}
}

// --- Dutiful Griffin -------------------------------------------------------

func TestStd52GriffinReturnsFromTheGraveyardForTwoEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	griffin := game.Card{InstanceID: uuid.New(), Name: "Dutiful Griffin", TypeLine: "Creature — Griffin", OracleID: std52GriffinOracle, Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(griffin)
	e1 := effectstest.B12Permanent(g, me.ID, "Enchantment One", "Enchantment")
	e2 := effectstest.B12Permanent(g, me.ID, "Enchantment Two", "Enchantment")
	e3 := effectstest.B12Permanent(g, me.ID, "Enchantment Three", "Enchantment")
	std52ToMain(t, g)
	hand := effectstest.HandSize(me)

	effectstest.B06AddMana(me, "C", "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, griffin.InstanceID, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{e1}}); err == nil {
		t.Fatal("one enchantment paid a two-enchantment cost")
	}
	me.ManaPool = nil
	effectstest.B06AddMana(me, "C", "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, griffin.InstanceID, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{e1, e2}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(e1) || g.Battlefield.Contains(e2) || !g.Battlefield.Contains(e3) {
		t.Error("exactly the two chosen enchantments should have been sacrificed")
	}
	if got := effectstest.HandSize(me); got != hand+1 || !me.Hand.Contains(griffin.InstanceID) {
		t.Errorf("hand %d (Griffin in hand: %v), want %d with the Griffin", got, me.Hand.Contains(griffin.InstanceID), hand+1)
	}
}
