package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aether Channeler — Creature {2}{U}, 2/1:
//
//	"When this creature enters, choose one —
//	 • Create a 1/1 white Bird creature token with flying.
//	 • Return another target nonland permanent to its owner's hand.
//	 • Draw a card."
func init() {
	channeler := WhenThisEnters("Aether Channeler — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	channeler.Modes = ChooseOne(
		ModeDoing("Create a 1/1 white Bird creature token with flying.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 white Bird with flying"), N: 1}.Apply(ctx)
		}),
		ModeDoing("Return another target nonland permanent to its owner's hand.",
			Another(TargetPermanent("another target nonland permanent", Nonland())),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return BounceToHand{Target: t.ID}.Apply(ctx)
			}),
		ModeDoing("Draw a card.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		}),
	)
	Register(Spec{
		OracleID:     "fb220f46-f8b8-4804-baa4-e7d50b4871f7",
		Name:         "Aether Channeler",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{channeler},
	})
}
