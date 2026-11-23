package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Assailant — Creature — Goblin Warrior {1}{R}, 2/2 (War of the Spark, #5520):
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
		OracleID:     "bc654150-b051-40bd-81f8-621abbb17e91",
		Name:         "Goblin Assailant",
		Completeness: effects.CompletenessFull,
	})
}
