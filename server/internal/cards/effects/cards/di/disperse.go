package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disperse — Instant {1}{U} (Scars of Mirrodin, #5577):
//
//	"Return target nonland permanent to its owner's hand."
//
// Boomerang Basics' bounce without the draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d8fe3f8-5027-4f9e-999e-e277caa96478",
		Name:         "Disperse",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.BounceTheTargetToHand(ctx.Game, item)
		},
	})
}
