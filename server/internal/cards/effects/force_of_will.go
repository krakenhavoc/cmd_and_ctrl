package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Force of Will — Instant {3}{U}{U}:
//
//	"You may pay 1 life and exile a blue card from your hand rather
//	 than pay this spell's mana cost.
//	 Counter target spell."
//
// The format-defining free counterspell, and the card that made the
// alternative-cost machinery grow a non-mana half. Two things worth
// knowing about how it behaves here:
//
//   - Both halves of the pitch are COSTS, validated together before
//     either is paid. A player at 1 life with no other blue card gets
//     a rejected cast, not a dead player and a spell still in hand.
//   - The exile happens with the Force already ON THE STACK (CR
//     601.2a before 601.2h). So countering it does not refund the
//     pitched card, which is the whole reason Force of Will is a fair
//     card rather than a free one — and the thing a resolution-time
//     implementation would get exactly backwards.
//
// Force of Will cannot pitch itself: CR 601.2a moved it to the stack
// before the cost was paid, so it is no longer in hand, and the
// engine rejects a cast that names it.
func init() {
	Register(Spec{
		OracleID: "956381ba-6d37-4a8a-846c-bad79222dbee",
		Name:     "Force of Will",
		Targets:  TargetSpell("target spell"),
		AlternativeCosts: []game.AlternativeCost{
			Pitch(
				"Pay 1 life and exile a blue card from your hand",
				1,
				CardInYourHand("a blue card from your hand", OfColor("U")),
				"a blue card from your hand",
			),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
