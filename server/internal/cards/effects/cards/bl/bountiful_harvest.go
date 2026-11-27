package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bountiful Harvest — Sorcery {4}{G} (Magic 2013, S82 OLD, #5568):
//
//	"You gain 1 life for each land you control."
//
// The lands are counted as the spell resolves (CR 608.2h), yours only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2dc94a25-4e35-4277-8eba-b8d0b293ff7b",
		Name:         "Bountiful Harvest",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GainLife{Amount: effects.CountYouControl(ctx, effects.Land())}.Apply(ctx)
		},
	})
}
