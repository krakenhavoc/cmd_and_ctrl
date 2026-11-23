package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestFandanielOpponentsWithNoNontokenCreatureLoseTwoPerSpell(t *testing.T) {
	g := effectstest.FandanielGame(t, 3)
	// Seat 2 has only a token; seats 1 and 3 have nothing.
	addWizardTokens(t, g, g.Seats[2].ID, 1)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 1; i <= 3; i++ {
		if got := g.Seats[i].Life; got != 40-6 {
			t.Errorf("seat %d life = %d, want 34 (2 x 3 instants and sorceries)", i, got)
		}
	}
	if g.Seats[0].Life != 40 {
		t.Errorf("Fandaniel's controller lost life: %d", g.Seats[0].Life)
	}
}

func TestFandanielOpponentMaySacrificeANontokenCreatureInsteadOfLosingLife(t *testing.T) {
	g := effectstest.FandanielGame(t, 2)
	opp := g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, opp.ID, 1) // sacrifice
	pick := effectstest.LatestChooseCardsFor(g, opp.ID)
	if pick == nil {
		t.Fatal("no sacrifice prompt for the opponent")
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the bear should have been sacrificed")
	}
	if opp.Life != 40 {
		t.Errorf("the sacrificing opponent lost life: %d", opp.Life)
	}
	// The others had nothing to sacrifice and lose 2 x 2.
	for i := 2; i <= 3; i++ {
		if got := g.Seats[i].Life; got != 36 {
			t.Errorf("seat %d life = %d, want 36", i, got)
		}
	}
}

func TestFandanielOpponentWhoDeclinesLosesLifeAndKeepsTheCreature(t *testing.T) {
	g := effectstest.FandanielGame(t, 2)
	opp := g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, opp.ID, 0) // don't sacrifice
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the bear should have stayed")
	}
	if opp.Life != 36 {
		t.Errorf("declining opponent life = %d, want 36", opp.Life)
	}
}

func TestFandanielOffersOnlyNontokenCreatures(t *testing.T) {
	g := effectstest.FandanielGame(t, 1)
	opp := g.Seats[1]
	addWizardTokens(t, g, opp.ID, 1)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, opp.ID, 1)
	pick := effectstest.LatestChooseCardsFor(g, opp.ID)
	if pick == nil {
		t.Fatal("no sacrifice prompt")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != bear {
		t.Errorf("sacrifice candidates = %v, want only the nontoken bear", pick.ChooseCards)
	}
}

func TestFandanielSurveilsWhenYouCastAnInstantOrSorcery(t *testing.T) {
	g := effectstest.FandanielGame(t, 0)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", effectstest.SpellControlDivinationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID)
	if c == nil {
		t.Fatal("casting a sorcery should have queued a surveil 1")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
}

func TestFandanielDoesNotSurveilOnACreatureSpell(t *testing.T) {
	g := effectstest.FandanielGame(t, 0)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID); c != nil {
		t.Error("a creature spell must not trigger Fandaniel's surveil")
	}
}
