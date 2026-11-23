package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whiptail Wurm — Creature — Wurm {6}{G}, 8/5 (Portal, #5630):
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
		OracleID:     "50fa6a63-e031-47cb-8fd5-a6c235203722",
		Name:         "Whiptail Wurm",
		Completeness: effects.CompletenessFull,
	})
}
