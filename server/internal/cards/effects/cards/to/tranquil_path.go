package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tranquil Path — Sorcery {4}{G} (Apocalypse, S82 OLD, #5614):
//
//	"Destroy all enchantments.
//	 Draw a card."
//
// Back to Nature's sweep, then the controller draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9595011-8ddb-403a-ba35-5a79d346557e",
		Name:         "Tranquil Path",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Sweep: game.Sweep{Matches: game.SweepEnchantments, How: game.SweepDestroy},
			Draws: 1,
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.DestroyAllMatching{Match: effects.Enchantment()}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
