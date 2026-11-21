package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Old Ghastbark — Creature — Treefolk Warrior {3}{G/W}{G/W}, 3/6
// (Shadowmoor, #5591):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// alaborn_trooper.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5207cb5-7519-4802-85c5-ac8c4313693e",
		Name:         "Old Ghastbark",
		Completeness: effects.CompletenessFull,
	})
}
