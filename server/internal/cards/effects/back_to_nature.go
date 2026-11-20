package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Back to Nature — Instant {1}{G} (Magic 2011, #5579):
//
//	"Destroy all enchantments."
//
// Cleanfall's sweep: every enchantment on the battlefield, its
// controller's own included, is destroyed at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b8d8ece6-397e-4bc9-b86c-7caa1d675d0f",
		Name:         "Back to Nature",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepEnchantments, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: Enchantment()}.Apply(ctx)
		},
	})
}
