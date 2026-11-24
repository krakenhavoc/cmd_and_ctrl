package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const glacianPowerstoneEngineerOracle = "e96ce18e-b002-4b5b-9560-c1634320e164"

func TestGlacianLooksAtXCardsTakesOneAndMillsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	glacian := effectstest.B12Push(g, me.ID, "Glacian, Powerstone Engineer", "Legendary Creature — Human Artificer", glacianPowerstoneEngineerOracle, 3, 6)
	effectstest.Unsick(g, glacian)
	one := effectstest.B21Push(g, me.ID, "Rock One", "Artifact", "", 0, 0)
	two := effectstest.B21Push(g, me.ID, "Rock Two", "Artifact", "", 0, 0)
	idle := effectstest.B21Push(g, me.ID, "Idle Rock", "Artifact", "", 0, 0)
	effectstest.SeedLibrary(me, "Z", "Y", "X", "W", "V")
	top := effectstest.SeedLibraryTop(me, "Top A", "Sorcery", "{1}")
	second := effectstest.SeedLibraryTop(me, "Top B", "Sorcery", "{1}")
	_ = second
	hand, yard, lib := me.Hand.Size(), me.Graveyard.Size(), me.Library.Size()

	effectstest.B16Activate(t, g, me.ID, glacian, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, TapIDs: []uuid.UUID{one, two}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCards(g, me.ID) == nil {
		t.Fatal("no \"put one of those cards into your hand\" prompt")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, top)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) || me.Hand.Size() != hand+1 {
		t.Errorf("the chosen card is not the only one in hand (hand %d, was %d)", me.Hand.Size(), hand)
	}
	if me.Graveyard.Size() != yard+1 || me.Library.Size() != lib-2 {
		t.Errorf("the rest did not go to the graveyard: graveyard %d (was %d), library %d (was %d)", me.Graveyard.Size(), yard, me.Library.Size(), lib)
	}
	if !effectstest.B16Tapped(t, g, glacian) || !effectstest.B16Tapped(t, g, one) || !effectstest.B16Tapped(t, g, two) {
		t.Error("Glacian and the two named artifacts were not tapped for the cost")
	}
	if effectstest.B16Tapped(t, g, idle) {
		t.Error("an artifact that was not named was tapped")
	}
}

func TestGlacianCannotTapItselfForTheArtifactCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	glacian := effectstest.B12Push(g, me.ID, "Glacian, Powerstone Engineer", "Legendary Artifact Creature — Human Artificer", glacianPowerstoneEngineerOracle, 3, 6)
	effectstest.Unsick(g, glacian)
	err := g.ActivateCatalogAbility(me.ID, glacian, 0, game.ActivateAbilityParams{Strict: true, XValue: 1, TapIDs: []uuid.UUID{glacian}})
	if err == nil {
		t.Error("Glacian paid X by tapping himself as well as his own {T}")
	}
}
