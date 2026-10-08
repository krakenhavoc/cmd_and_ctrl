package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// betor_prompts_test.go — #2523 / #2524: the resolution-time card picks
// of Colossal Grave-Reaver and Eerie Ultimatum are answerable by the
// legal-move enumerator (what a bot picks from), and never offer an
// answer the card forbids.

// chooseCardsAnswers returns the card-id sets the enumerator offers the
// seat for its open prompt.
func chooseCardsAnswers(t *testing.T, g *game.Game, seat uuid.UUID) [][]uuid.UUID {
	t.Helper()
	var out [][]uuid.UUID
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Kind != legal.KindChoice {
			continue
		}
		var p struct {
			CardIDs []uuid.UUID `json:"card_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			continue
		}
		out = append(out, p.CardIDs)
	}
	return out
}

func TestEerieUltimatumEnumeratorOffersOnlyDifferentNamedSets(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bearA := seedGraveyardCard(t, g, "Bear", "Creature — Bear", "")
	bearB := seedGraveyardCard(t, g, "Bear", "Creature — Bear", "")
	land := seedGraveyardCard(t, g, "Forest", "Basic Land — Forest", "")
	castCatalogSpell(t, g, "Eerie Ultimatum", "Sorcery", b14EerieUltimatumOracle, nil)
	passPriorityAroundTable(t, g)

	sets := chooseCardsAnswers(t, g, me.ID)
	if len(sets) == 0 {
		t.Fatal("the enumerator offers a seat owing the prompt nothing")
	}
	var none, pair bool
	for _, s := range sets {
		has := map[uuid.UUID]bool{}
		for _, id := range s {
			has[id] = true
		}
		if has[bearA] && has[bearB] {
			t.Errorf("offered a set with two Bears: %v", s)
		}
		if len(s) == 0 {
			none = true
		}
		if len(s) == 2 && has[land] && (has[bearA] || has[bearB]) {
			pair = true
		}
	}
	if !none {
		t.Error("\"any number\" includes none: the empty answer is offered")
	}
	if !pair {
		t.Error("a Bear and the Forest is a legal set and is offered")
	}
}

func TestColossalGraveReaverEnumeratorOffersExactlyOneCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Small Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Big Wurm", TypeLine: "Creature — Wurm", ManaCost: "{5}{G}", Power: 6, Toughness: 6},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Deep Card", TypeLine: "Instant"},
	)
	castCatalogSpell(t, g, "Colossal Grave-Reaver", "Creature — Dragon", b17ColossalGraveReaverOracle, nil)
	passPriorityAroundTable(t, g)

	sets := chooseCardsAnswers(t, g, me.ID)
	if len(sets) != 2 {
		t.Fatalf("one answer per milled creature card, got %d: %v", len(sets), sets)
	}
	for _, s := range sets {
		if len(s) != 1 {
			t.Errorf("an answer must name exactly one card, got %v", s)
		}
	}
}
