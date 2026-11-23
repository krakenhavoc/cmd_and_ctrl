package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Muck Rats — Creature — Rat {B}, 1/1 (Portal, #5630):
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
		OracleID:     "bca13a12-6723-4a5e-8f1b-21646a8b3e7e",
		Name:         "Muck Rats",
		Completeness: effects.CompletenessFull,
	})
}
