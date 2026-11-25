package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Natural Obsolescence — {1}{G} Instant (Aether Revolt, #5536):
//
//	"Put target artifact on the bottom of its owner's library."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fbafbe3a-3980-4457-a77b-3f72df7b3329",
		Name:         "Natural Obsolescence",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact", effects.Artifact()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SocB37PutTargetGraveyardCardOnBottom(ctx.Game, item)
		},
	})
}
