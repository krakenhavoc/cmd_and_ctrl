package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Devastation — Sorcery {5}{R}{R} (Portal, #5630):
//
//	"Destroy all creatures and lands."
//
// Day of Judgment and Armageddon in one destruction: every creature and
// every land is destroyed at the same time.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "052838cb-dcf0-46f5-82b1-c3ed863b42b7",
		Name:         "Devastation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Or(effects.Creature(), effects.Land())}.Apply(ctx)
		},
	})
}
