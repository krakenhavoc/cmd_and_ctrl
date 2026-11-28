package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin War Cry — Sorcery {2}{R} (Portal Second Age, S82 older sets batch 735, #5625):
//
//	"Target opponent chooses a creature they control. Other creatures they control can't block this turn."
//
// The choice is the target opponent's, made at resolution; the other creatures are read live as blockers are declared (CR 611.2c), so one that player gets later this turn can't block either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6c8c660-2b06-4395-8c86-dd540760d9ae",
		Name:         "Goblin War Cry",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB735GoblinWarCry(ctx.Game, item)
		},
	})
}
