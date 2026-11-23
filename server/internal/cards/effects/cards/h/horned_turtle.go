package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Horned Turtle — Creature — Turtle {2}{U}, 1/4 (Portal, #5630):
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
		OracleID:     "7a4ae76b-6241-4fa8-b239-5ec941a0e7df",
		Name:         "Horned Turtle",
		Completeness: effects.CompletenessFull,
	})
}
