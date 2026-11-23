package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rest_in_peace_test.go — the proof for #931. Rest in Peace and
// Leyline of the Void are one sentence, "if a card would be put into a
// graveyard from anywhere, exile it instead", and the word that does
// the work is ANYWHERE: the card is only correct if every route into a
// graveyard opens the CR 614 window. The two that did not were a
// library search with a graveyard destination and surveil's graveyard
// leg, which is why these cards sat on #383's skip list until the
// routes were fixed.
//
// So the tests are one per route rather than one per card: search,
// surveil, mill, a discard and a creature dying. Each of them is a
// different mover in the engine and the same sentence on the card.

const (
	RestInPeaceOracle      = "087f9ad7-e74f-40e2-8102-1ed2925d0418"
	LeylineOfTheVoidOracle = "f4e32fc1-1b8d-441e-8e76-71f19f98e925"
)

// castRestInPeace casts it from the active seat's hand and lets it
// resolve, so its enters-the-battlefield sweep really fires.
func CastRestInPeace(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := CastCatalogSpell(t, g, "Rest in Peace", "Enchantment", RestInPeaceOracle, nil)
	PassPriorityAroundTable(t, g)
	return id
}
