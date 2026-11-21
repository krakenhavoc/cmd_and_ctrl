package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feral Maaka — Creature — Cat {1}{R}, 2/2 (Ravnica Allegiance, #5521):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// grizzly_bears.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d4556b82-f96e-4f88-a885-3e84afdb98a1",
		Name:         "Feral Maaka",
		Completeness: effects.CompletenessFull,
	})
}
