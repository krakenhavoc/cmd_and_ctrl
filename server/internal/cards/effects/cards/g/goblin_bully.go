package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Bully — Creature — Goblin {1}{R}, 2/1 (Portal, #5630):
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
		OracleID:     "56c5afdc-6777-45c7-8e2c-3cadecb95c5a",
		Name:         "Goblin Bully",
		Completeness: effects.CompletenessFull,
	})
}
