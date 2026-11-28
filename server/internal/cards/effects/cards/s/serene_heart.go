package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serene Heart — Instant {1}{G} (Mirage):
//
//	"Destroy all Auras."
//
// Every Aura on the battlefield, whoever controls it, is destroyed at once
// (CR 700.4); no regeneration clause. Declared as a partial enchantment sweep.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f38b03dd-a200-48a6-ab65-fb2ac47456df",
		Name:         "Serene Heart",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepEnchantments, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.And(effects.Enchantment(), effects.Subtype("Aura"))}.Apply(ctx)
		},
	})
}
