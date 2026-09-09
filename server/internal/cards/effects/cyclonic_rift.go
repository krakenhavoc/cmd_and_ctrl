package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cyclonic Rift — Instant {1}{U}:
//
//	"Return target nonland permanent you don't control to its owner's
//	 hand.
//	 Overload {6}{U} (You may cast this spell for its overload cost.
//	 If you do, change 'target' to 'each'.)"
//
// Sandbox simplification: OVERLOAD IS NOT IMPLEMENTED, which in this
// card's case removes the reason people play it — the {6}{U} one-sided
// board wipe. Only the two-mana single-bounce mode exists here. Same
// blocker as Vandalblast: overload is an alternative cost that also
// rewrites the target clause.
//
// Registered anyway because the base mode is a legal, useful play and
// the card is common enough in decklists that leaving it non-catalog
// would drop it to manual sandbox resolution. The deferral is loud so
// nobody mistakes a bounce for the wipe they expected.
func init() {
	Register(Spec{
		OracleID: "d75b9c82-1b49-4c3e-a1b5-aeef57d6644b",
		Name:     "Cyclonic Rift",
		Targets: TargetPermanent("target nonland permanent you don't control",
			And(Nonland(), OpponentControls())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BounceToHand{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
