package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// solemn_simulacrum_test.go — the clauses of Solemn Simulacrum that
// cards_test.go's ETB and dies tests did not pin, written when the
// card was read against its oracle text and marked CompletenessFull
// (#1306).

// TestSolemnSimulacrumOffersOnlyBasicsAndTheyEnterTapped — "a basic
// land card … onto the battlefield tapped".
func TestSolemnSimulacrumOffersOnlyBasicsAndTheyEnterTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Breeding Pool", TypeLine: "Land — Forest Island"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Snow-Covered Island", TypeLine: "Basic Snow Land — Island"},
	)
	castCatalogSpell(t, g, "Solemn Simulacrum", "Artifact Creature — Golem", b06SolemnSimulacrumOracle, nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two basics match, so the search asks which")
	}
	if searchOptionNamed(g, c, "Breeding Pool") != uuid.Nil {
		t.Error("a nonbasic land is offered")
	}
	// A snow basic is a basic land card.
	answerSearchNamed(t, g, me.ID, "Snow-Covered Island")
	island := findBattlefieldByName(g, "Snow-Covered Island")
	if island == uuid.Nil {
		t.Fatal("the chosen basic did not enter")
	}
	if c, _ := battlefieldCard(g, island); !c.Tapped {
		t.Error("the basic entered untapped; the card says tapped")
	}
}

// TestSolemnSimulacrumDeclinedSearchTakesNothing — "you MAY search":
// declining leaves the library alone.
func TestSolemnSimulacrumDeclinedSearchTakesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	castCatalogSpell(t, g, "Solemn Simulacrum", "Artifact Creature — Golem", b06SolemnSimulacrumOracle, nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forest) || !me.Library.Contains(forest) {
		t.Error("a declined search still took the Forest")
	}
}
