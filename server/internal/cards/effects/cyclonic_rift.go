package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cyclonic Rift — Instant {1}{U}:
//
//	"Return target nonland permanent you don't control to its owner's
//	 hand.
//	 Overload {6}{U} (You may cast this spell for its overload cost.
//	 If you do, change 'target' in its text to 'each.')"
//
// S22: overload now works, which restores the reason people play the
// card — the {6}{U} one-sided board wipe at instant speed. Before
// this the catalog shipped it as a two-mana single bounce, which is a
// different and much worse card.
//
// "You don't control" is what makes it one-sided, and it is enforced
// in the sweep as well as in the targeted mode: your own board is
// never touched.
func init() {
	Register(Spec{
		OracleID: "d75b9c82-1b49-4c3e-a1b5-aeef57d6644b",
		Name:     "Cyclonic Rift",
		Targets: TargetPermanent("target nonland permanent you don't control",
			And(Nonland(), OpponentControls())),
		AlternativeCosts: []game.AlternativeCost{
			Overload("{6}{U}"),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.PaidAltCost("overload") {
				// Snapshot first — BounceToHand mutates the
				// battlefield slice, and a leaves-the-battlefield
				// trigger could add to it mid-sweep.
				var doomed []uuid.UUID
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if !c.IsLand() && c.Controller != item.Controller {
						doomed = append(doomed, c.InstanceID)
					}
				}
				for _, id := range doomed {
					if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BounceToHand{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
