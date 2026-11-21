package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rona's Vortex — Instant {U} (Dominaria United):
//
//	"Kicker {2}{B}
//	 Return target creature or planeswalker you don't control to its owner's
//	 hand. If this spell was kicked, put that permanent on the bottom of its
//	 owner's library instead."
//
// One target, two destinations (Zuko's Conviction's shape): unkicked it is a
// bounce, kicked it goes on the bottom of its owner's library. A target that
// is gone or no longer a creature or planeswalker you don't control by
// resolution is illegal and nothing happens (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "21989fad-9672-46d4-b7d5-101906e7aed2",
		Name:          "Rona's Vortex",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{2}{B}")},
		Targets: effects.TargetPermanent("target creature or planeswalker you don't control",
			effects.Or(effects.Creature(), effects.Planeswalker()), effects.Not(effects.YouControl())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if ctx.WasKicked() {
				return ctx.Game.TuckToLibraryForEffect(target, true)
			}
			return effects.BounceToHand{Target: target}.Apply(ctx)
		},
	})
}
