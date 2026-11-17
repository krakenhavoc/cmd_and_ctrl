package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "21989fad-9672-46d4-b7d5-101906e7aed2",
		Name:          "Rona's Vortex",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{2}{B}")},
		Targets: TargetPermanent("target creature or planeswalker you don't control",
			Or(Creature(), Planeswalker()), Not(YouControl())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if ctx.WasKicked() {
				return ctx.Game.TuckToLibraryForEffect(target, true)
			}
			return BounceToHand{Target: target}.Apply(ctx)
		},
	})
}
