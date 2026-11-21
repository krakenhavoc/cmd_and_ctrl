package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Devoted Hero — Creature — Elf Soldier {W}, 1/2 (Portal, #5630):
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
		OracleID:     "fff6f2d0-22b5-41d0-99dc-bc817c3bd166",
		Name:         "Devoted Hero",
		Completeness: effects.CompletenessFull,
	})
}
