package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clip Wings — Instant {1}{G} (Shadows over Innistrad, S82 OLD):
//
//	"Each opponent sacrifices a creature of their choice with flying."
//
// Each opponent picks their own flier; one with none sacrifices nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cefa28e2-6a95-4aca-86f4-b0b741f61939",
		Name:         "Clip Wings",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.EachPlayerSacrifices{
				ExceptController: true,
				Match:            effects.And(effects.Creature(), effects.HasKeyword("flying")),
				Label:            "a creature with flying",
			}.Apply(ctx)
		},
	})
}
