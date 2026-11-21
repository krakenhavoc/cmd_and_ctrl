package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Culling Sun — Sorcery {2}{W}{W}{B} (Guildpact, #5599):
//
//	"Destroy each creature with mana value 3 or less."
//
// Austere Command's third bullet as a whole spell: DestroyAllMatching
// over creatures whose mana value (ManaValueLE, the layered value of the
// permanent) is 3 or less. A token copy has its original's mana value
// and a plain token has 0 (CR 202.3), so both are destroyed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e41d9b3a-0712-4174-8c5c-224065a63b16",
		Name:         "Culling Sun",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreaturesManaValue3OrLess, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.ManaValueLE(3))}.Apply(ctx)
		},
	})
}
