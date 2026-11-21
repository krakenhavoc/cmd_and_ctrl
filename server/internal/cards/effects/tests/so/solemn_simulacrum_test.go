package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// solemn_simulacrum_test.go — the clauses of Solemn Simulacrum that
// cards_test.go's ETB and dies tests did not pin, written when the
// card was read against its oracle text and marked CompletenessFull
// (#1306).

// TestSolemnSimulacrumOffersOnlyBasicsAndTheyEnterTapped — "a basic
// land card … onto the battlefield tapped".
func TestSolemnSimulacrumOffersOnlyBasicsAndTheyEnterTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Breeding Pool", TypeLine: "Land — Forest Island"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Snow-Covered Island", TypeLine: "Basic Snow Land — Island"},
	)
	effectstest.CastCatalogSpell(t, g, "Solemn Simulacrum", "Artifact Creature — Golem", effectstest.B06SolemnSimulacrumOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two basics match, so the search asks which")
	}
	if effectstest.SearchOptionNamed(g, c, "Breeding Pool") != uuid.Nil {
		t.Error("a nonbasic land is offered")
	}
	// A snow basic is a basic land card.
	effectstest.AnswerSearchNamed(t, g, me.ID, "Snow-Covered Island")
	island := effectstest.FindBattlefieldByName(g, "Snow-Covered Island")
	if island == uuid.Nil {
		t.Fatal("the chosen basic did not enter")
	}
	if c, _ := effectstest.BattlefieldCard(g, island); !c.Tapped {
		t.Error("the basic entered untapped; the card says tapped")
	}
}

// TestSolemnSimulacrumDeclinedSearchTakesNothing — "you MAY search":
// declining leaves the library alone.
func TestSolemnSimulacrumDeclinedSearchTakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.CastCatalogSpell(t, g, "Solemn Simulacrum", "Artifact Creature — Golem", effectstest.B06SolemnSimulacrumOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forest) || !me.Library.Contains(forest) {
		t.Error("a declined search still took the Forest")
	}
}
