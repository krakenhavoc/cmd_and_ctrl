package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight Errant — Creature — Human Knight {1}{W}, 2/2 (Portal, #5630):
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
		OracleID:     "920cc1a7-ac6b-44c0-ade7-347f05bba0f2",
		Name:         "Knight Errant",
		Completeness: effects.CompletenessFull,
	})
}
