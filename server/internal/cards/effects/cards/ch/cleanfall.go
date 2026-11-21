package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cleanfall — Sorcery — Arcane {2}{W} (Champions of Kamigawa, #5603):
//
//	"Destroy all enchantments."
//
// Every enchantment on the battlefield, its caster's included; an
// indestructible one stays (CR 702.12b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "826b1c93-4e5d-4436-8d59-2d7d970f609f",
		Name:         "Cleanfall",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepEnchantments, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Enchantment()}.Apply(ctx)
		},
	})
}
