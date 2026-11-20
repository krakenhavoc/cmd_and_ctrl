package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tempest of Light — Instant {2}{W} (Mirrodin, #5606):
//
//	"Destroy all enchantments."
//
// Cleanfall's sweep: every enchantment on the battlefield, whoever
// controls it, is destroyed at once (CR 701.8a); an indestructible
// one survives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9399ed27-31c7-4d6a-99eb-644b74e67344",
		Name:         "Tempest of Light",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepEnchantments, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: Enchantment()}.Apply(ctx)
		},
	})
}
