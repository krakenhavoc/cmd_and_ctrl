package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Raider — Creature — Goblin Warrior {1}{R}, 2/2 (Portal Second Age,
// #5625):
//
//	"This creature can't block."
//
// Jungle Lion's block ban: a RestrictSelf static the declare-blockers
// check reads (CR 509.1b). It can still attack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f11c830-fced-40db-8c02-b3c54f3e4372",
		Name:         "Goblin Raider",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
	})
}
