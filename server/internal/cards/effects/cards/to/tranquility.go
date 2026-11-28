package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tranquility — Sorcery {2}{G} (Invasion, #5616):
//
//	"Destroy all enchantments."
//
// Back to Nature's sweep: every enchantment on the battlefield, its
// controller's own included, is destroyed at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f671e3c3-cd59-4d06-a1af-5d04892cf74d",
		Name:         "Tranquility",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepEnchantments, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Enchantment()}.Apply(ctx)
		},
	})
}
