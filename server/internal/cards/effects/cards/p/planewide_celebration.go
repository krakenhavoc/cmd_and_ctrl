package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Planewide Celebration — Sorcery {5}{G}{G} (War of the Spark, #5520):
//
//	"Choose four. You may choose the same mode more than once.
//	 • Create a 2/2 Citizen creature token that's all colors.
//	 • Return target permanent card from your graveyard to your hand.
//	 • Proliferate.
//	 • You gain 4 life."
//
// Righteous Confluence's ChooseNRepeating (CR 700.2d): each occurrence runs
// once in printed order, and each chosen return gets its own target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af23d348-adc9-451e-af08-bf9b1986c591",
		Name:         "Planewide Celebration",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseNRepeating("Choose four (you may choose the same mode more than once)", 4, 4,
			effects.ModeDoing("Create a 2/2 Citizen creature token that's all colors.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("2/2 white and blue and black and red and green Citizen"), N: 1}.Apply(ctx)
				}),
			effects.ModeDoing("Return target permanent card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target permanent card in your graveyard", effects.YouOwn(), effects.Permanent()),
				effects.S82OldB170ReturnModeTargetToHand),
			effects.ModeDoing("Proliferate.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.Proliferate{}.Apply(ctx)
				}),
			effects.ModeDoing("You gain 4 life.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.GainLife{Player: item.Controller, Amount: 4}.Apply(ctx)
				}),
		),
	})
}
