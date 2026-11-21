package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Righteous Confluence — Sorcery {3}{W}{W} (Commander 2015, #5546):
//
//	"Choose three. You may choose the same mode more than once.
//	 • Create a 2/2 white Knight creature token with vigilance.
//	 • Exile target enchantment.
//	 • You gain 5 life."
//
// Mystic Confluence's ChooseNRepeating (CR 700.2d): each occurrence runs
// once in printed order, and each chosen exile gets its own target group.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee28937b-7ead-45ad-8448-35c51a09fb32",
		Name:         "Righteous Confluence",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseNRepeating("Choose three (you may choose the same mode more than once)", 3, 3,
			effects.ModeDoing("Create a 2/2 white Knight creature token with vigilance.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("2/2 white Knight with vigilance"), N: 1}.Apply(ctx)
				}),
			effects.ModeDoing("Exile target enchantment.",
				effects.TargetPermanent("target enchantment", effects.Enchantment()),
				effects.ExileTheModesTarget),
			effects.ModeDoing("You gain 5 life.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.GainLife{Player: item.Controller, Amount: 5}.Apply(ctx)
				}),
		),
	})
}
