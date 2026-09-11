package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cyclonic Rift — Instant {1}{U}:
//
//	"Return target nonland permanent you don't control to its owner's
//	 hand.
//	 Overload {6}{U} (You may cast this spell for its overload cost.
//	 If you do, change "target" in its text to "each.")"
//
// S22 made overload work, which restored the reason people play the
// card — the {6}{U} one-sided board wipe at instant speed. Before
// that the catalog shipped it as a two-mana single bounce, which is a
// different and much worse card.
//
// S23: the overloaded sweep is BounceAllMatching over the exact
// predicate the targeted mode is built from, And(Nonland(),
// OpponentControls()). "You don't control" is what makes the card
// one-sided and it is enforced identically in both modes — your own
// board is never touched.
func init() {
	nonlandYouDontControl := And(Nonland(), OpponentControls())
	Register(Spec{
		OracleID: "d75b9c82-1b49-4c3e-a1b5-aeef57d6644b",
		Name:     "Cyclonic Rift",
		Targets: TargetPermanent("target nonland permanent you don't control",
			nonlandYouDontControl),
		AlternativeCosts: []game.AlternativeCost{
			Overload("{6}{U}"),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.PaidAltCost("overload") {
				return BounceAllMatching{Match: nonlandYouDontControl}.Apply(ctx)
			}
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BounceToHand{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
