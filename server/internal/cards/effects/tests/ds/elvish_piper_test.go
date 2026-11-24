package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const elvishPiperOracle = "872ef617-7cdb-4a7a-85f5-efc9e12bece5"

func TestElvishPiperIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(elvishPiperOracle)
	if !ok || spec.Name != "Elvish Piper" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Elvish Piper spec = %+v, registered %v", spec, ok)
	}
}

func elvishPiperHand(me *game.Player) (fatty, shock uuid.UUID) {
	fatty, shock = uuid.New(), uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: fatty, Name: "Big Fatty", TypeLine: "Creature — Wurm", ManaCost: "{8}",
		Power: 8, Toughness: 8, Owner: me.ID, Controller: me.ID})
	me.Hand.PushTop(game.Card{InstanceID: shock, Name: "Shock", TypeLine: "Instant", ManaCost: "{R}",
		Owner: me.ID, Controller: me.ID})
	return fatty, shock
}

func TestElvishPiperPutsACreatureFromHandWithoutCastingIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	piper := effectstest.PushCatalogPermanent(g, me.ID, "Elvish Piper", "Creature — Elf Shaman", elvishPiperOracle, false)
	fatty, shock := elvishPiperHand(me)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G")
	effectstest.B16Activate(t, g, me.ID, piper, 0, game.ActivateAbilityParams{})
	if !effectstest.B16Tapped(t, g, piper) {
		t.Error("the Piper did not tap to pay")
	}
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("the put-from-hand prompt did not open")
	}
	if !effectstest.HasID(c.ChooseCards, fatty) || effectstest.HasID(c.ChooseCards, shock) {
		t.Errorf("offered %v: want the creature card and not the instant", c.ChooseCards)
	}
	if c.ChooseMin != 0 {
		t.Error("\"you may put\": declining must be a legal answer")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{fatty}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Battlefield.Contains(fatty) || me.Hand.Contains(fatty) {
		t.Error("the creature did not arrive on the battlefield")
	}
}

func TestElvishPiperMayDeclineAndNeedsGreenAndAnUntappedPiper(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	piper := effectstest.PushCatalogPermanent(g, me.ID, "Elvish Piper", "Creature — Elf Shaman", elvishPiperOracle, false)
	fatty, _ := elvishPiperHand(me)
	effectstest.AdvanceToMain(t, g)

	strict := game.ActivateAbilityParams{Strict: true}
	if err := g.ActivateCatalogAbility(me.ID, piper, 0, strict); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, piper, 0, strict); err == nil {
		t.Fatal("activated with colorless mana for {G}")
	}
	effectstest.B06AddMana(me, "G")
	effectstest.B16Activate(t, g, me.ID, piper, 0, game.ActivateAbilityParams{})
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	if g.Battlefield.Contains(fatty) {
		t.Error("declined but the creature arrived")
	}
	effectstest.B06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, piper, 0, strict); err == nil {
		t.Error("activated a second time while tapped")
	}
}
