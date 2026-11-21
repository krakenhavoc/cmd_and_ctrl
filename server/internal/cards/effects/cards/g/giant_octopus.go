package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Giant Octopus — Creature — Octopus {3}{U}, 3/3 (Portal, #5630):
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
		OracleID:     "c181d2a4-5959-4409-9bd3-ecedf8ec9516",
		Name:         "Giant Octopus",
		Completeness: effects.CompletenessFull,
	})
}
