package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Back to Nature — Instant {1}{G} (Magic 2011, #5579):
//
//	"Destroy all enchantments."
//
// Cleanfall's sweep: every enchantment on the battlefield, its
// controller's own included, is destroyed at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8d8ece6-397e-4bc9-b86c-7caa1d675d0f",
		Name:         "Back to Nature",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepEnchantments, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Enchantment()}.Apply(ctx)
		},
	})
}
