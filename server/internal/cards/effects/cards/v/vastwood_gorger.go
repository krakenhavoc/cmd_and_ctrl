package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vastwood Gorger — Creature — Wurm {5}{G}, 5/6 (Zendikar, #5583):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// ordinary_bear.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a17a86a3-9f9c-4e09-93e6-e543a70733bc",
		Name:         "Vastwood Gorger",
		Completeness: effects.CompletenessFull,
	})
}
