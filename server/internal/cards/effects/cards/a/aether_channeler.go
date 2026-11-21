package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aether Channeler — Creature {2}{U}, 2/1:
//
//	"When this creature enters, choose one —
//	 • Create a 1/1 white Bird creature token with flying.
//	 • Return another target nonland permanent to its owner's hand.
//	 • Draw a card."
func init() {
	channeler := effects.WhenThisEnters("Aether Channeler — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	channeler.Modes = effects.ChooseOne(
		effects.ModeDoing("Create a 1/1 white Bird creature token with flying.", nil, func(item *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 white Bird with flying"), N: 1}.Apply(ctx)
		}),
		effects.ModeDoing("Return another target nonland permanent to its owner's hand.",
			effects.Another(effects.TargetPermanent("another target nonland permanent", effects.Nonland())),
			effects.BounceTheModesTarget),
		effects.ModeDoing("Draw a card.", nil, func(item *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		}),
	)
	effects.Register(effects.Spec{
		OracleID:     "fb220f46-f8b8-4804-baa4-e7d50b4871f7",
		Name:         "Aether Channeler",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{channeler},
	})
}
