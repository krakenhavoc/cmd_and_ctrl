package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Redwood Treefolk — Creature — Treefolk {4}{G}, 3/6 (Portal, #5630):
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
		OracleID:     "2cdebbc2-2e0b-4d26-bb7f-4736f0318ca7",
		Name:         "Redwood Treefolk",
		Completeness: effects.CompletenessFull,
	})
}
