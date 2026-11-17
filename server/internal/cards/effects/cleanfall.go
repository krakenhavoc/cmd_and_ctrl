package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cleanfall — Sorcery — Arcane {2}{W} (Champions of Kamigawa, #5603):
//
//	"Destroy all enchantments."
//
// Every enchantment on the battlefield, its caster's included; an
// indestructible one stays (CR 702.12b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "826b1c93-4e5d-4436-8d59-2d7d970f609f",
		Name:         "Cleanfall",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepEnchantments, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: Enchantment()}.Apply(ctx)
		},
	})
}
