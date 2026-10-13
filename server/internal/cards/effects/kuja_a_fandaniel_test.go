package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kuja_a_fandaniel_test.go — Fandaniel, Telophoroi Ascian (#3213).

func fandanielGame(t *testing.T, spellsInGraveyard int) *game.Game {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushDiesCreatureForTest(g, me.ID, "Fandaniel, Telophoroi Ascian", fandanielOracle, "Legendary Creature — Elder Wizard", 4, 5)
	for i := 0; i < spellsInGraveyard; i++ {
		typ := "Instant"
		if i%2 == 1 {
			typ = "Sorcery"
		}
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spent Spell", TypeLine: typ, Owner: me.ID, Controller: me.ID})
	}
	// A creature card in the graveyard does not count.
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	return g
}

func TestFandanielOpponentsWithNoNontokenCreatureLoseTwoPerSpell(t *testing.T) {
	g := fandanielGame(t, 3)
	// Seat 2 has only a token; seats 1 and 3 have nothing.
	addWizardTokens(t, g, g.Seats[2].ID, 1)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
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
	g := fandanielGame(t, 2)
	opp := g.Seats[1]
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, opp.ID, 1) // sacrifice
	pick := latestChooseCardsFor(g, opp.ID)
	if pick == nil {
		t.Fatal("no sacrifice prompt for the opponent")
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
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
	g := fandanielGame(t, 2)
	opp := g.Seats[1]
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, opp.ID, 0) // don't sacrifice
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the bear should have stayed")
	}
	if opp.Life != 36 {
		t.Errorf("declining opponent life = %d, want 36", opp.Life)
	}
}

func TestFandanielOffersOnlyNontokenCreatures(t *testing.T) {
	g := fandanielGame(t, 1)
	opp := g.Seats[1]
	addWizardTokens(t, g, opp.ID, 1)
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, opp.ID, 1)
	pick := latestChooseCardsFor(g, opp.ID)
	if pick == nil {
		t.Fatal("no sacrifice prompt")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != bear {
		t.Errorf("sacrifice candidates = %v, want only the nontoken bear", pick.ChooseCards)
	}
}

func TestFandanielSurveilsWhenYouCastAnInstantOrSorcery(t *testing.T) {
	g := fandanielGame(t, 0)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Divination", "Sorcery", spellControlDivinationOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID)
	if c == nil {
		t.Fatal("casting a sorcery should have queued a surveil 1")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
}

func TestFandanielDoesNotSurveilOnACreatureSpell(t *testing.T) {
	g := fandanielGame(t, 0)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if c := latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID); c != nil {
		t.Error("a creature spell must not trigger Fandaniel's surveil")
	}
}
