package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cowl Prowler — Creature — Wurm {4}{G}{G}, 6/6 (Kaladesh, #5539):
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
		OracleID:     "67419470-3418-466f-b3c1-4ded1d0ee34b",
		Name:         "Cowl Prowler",
		Completeness: effects.CompletenessFull,
	})
}
