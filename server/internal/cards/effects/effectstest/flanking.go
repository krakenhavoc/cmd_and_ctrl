package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flankingImport is a card as the deck importer hands it to a game: its
// Scryfall record, with the printed text from the coverage oracle fixture
// when the card has one, run through deck.ToGameCard. An empty oracle is a
// card the catalog does not know.
func FlankingImport(t *testing.T, oracle, name, typeLine, cost string, power, toughness string, keywords ...string) game.Card {
	t.Helper()
	id := uuid.New()
	text := ""
	if oracle != "" {
		id = uuid.MustParse(oracle)
		if text = SpellText(t, oracle); text == "" {
			t.Fatalf("%s has no oracle fixture", name)
		}
	} else if _, ok := effects.Lookup(id.String()); ok {
		t.Fatalf("%s must not be in the catalog", name)
	}
	return deck.ToGameCard(cards.Card{
		ID: uuid.New(), OracleID: id, Name: name, Layout: "normal",
		TypeLine: typeLine, ManaCost: cost, Power: power, Toughness: toughness,
		Keywords: keywords, OracleText: text,
	}, false)
}

func FlankingBears(t *testing.T) game.Card {
	return FlankingImport(t, "", "Grizzly Bears", "Creature — Bear", "{1}{G}", "2", "2")
}

// flankingBlocks has the active seat's `attacker` attack `defender`, has
// each of `blockers` block it, locks the blocks in and resolves what
// triggered, stopping in the declare blockers step.
func FlankingBlocks(t *testing.T, g *game.Game, defender, attacker uuid.UUID, blockers ...uuid.UUID) {
	t.Helper()
	// #5784: the attack is locked in and its triggers resolved (Barbed
	// Foliage, Telim'Tor) before the blocks.
	DeclareAttack(t, g, defender, attacker)
	PassPriorityAroundTable(t, g)
	AdvanceTo(t, g, game.StepDeclareBlockers)
	for _, b := range blockers {
		if err := g.DeclareBlocker(b, attacker); err != nil {
			t.Fatal(err)
		}
	}
	if !FinishBlockDeclarations(t, g) {
		t.Fatal("no block declaration was left to finish")
	}
	PassPriorityAroundTable(t, g)
	if g.Turn.Step != game.StepDeclareBlockers {
		t.Fatalf("combat left the declare blockers step (now %v)", g.Turn.Step)
	}
}
