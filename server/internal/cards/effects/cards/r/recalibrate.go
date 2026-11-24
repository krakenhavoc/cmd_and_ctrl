package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Recalibrate — Instant {1}{U} (Modern Horizons 2, #5501):
//
//	"Return target creature to its owner's hand. If you've discarded a card
//	 this turn, draw a card."
//
// The bounce is Just the Wind's. The rider is read as the spell resolves
// against the caster's discard tally for the turn (CR 701.9a); with its one
// target gone the spell does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6331b3a-81a8-48e8-8a28-237d381f9062",
		Name:         "Recalibrate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.BounceTheTarget(item, ctx); err != nil {
				return err
			}
			if !effects.OldB030DiscardedACardThisTurn(ctx.Game, ctx.Controller()) {
				return nil
			}
			return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
