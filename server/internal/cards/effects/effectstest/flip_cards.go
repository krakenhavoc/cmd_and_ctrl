package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// flip_cards_test.go — the three proof cards of ADR 0151 (#1892), each
// down the real import road, so the per-half keywords, the colour and
// cost of the flipped half and the "#1" catalog key are the importer's.
// The engine verb is pinned in game/flip_test.go.

// flipRow is a flip card's Scryfall record: the colour and cost at the
// top level and on the top half, the flipped half with neither.
func FlipRow(oracleID, cost string, colors, keywords []string, top, flipped cards.CardFace) cards.Card {
	top.ManaCost = cost
	return cards.Card{
		ID:            uuid.New(),
		OracleID:      uuid.MustParse(oracleID),
		Name:          top.Name + " // " + flipped.Name,
		Layout:        "flip",
		TypeLine:      top.TypeLine + " // " + flipped.TypeLine,
		ManaCost:      cost,
		Colors:        colors,
		ColorIdentity: colors,
		Keywords:      keywords,
		CardFaces:     []cards.CardFace{top, flipped},
	}
}
