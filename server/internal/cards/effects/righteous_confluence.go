package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ee28937b-7ead-45ad-8448-35c51a09fb32",
		Name:         "Righteous Confluence",
		Completeness: CompletenessFull,
		Modes: ChooseNRepeating("Choose three (you may choose the same mode more than once)", 3, 3,
			ModeDoing("Create a 2/2 white Knight creature token with vigilance.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return CreateToken{Controller: ctx.Controller(), Template: TokenCard("2/2 white Knight with vigilance"), N: 1}.Apply(ctx)
				}),
			ModeDoing("Exile target enchantment.",
				TargetPermanent("target enchantment", Enchantment()),
				ExileTheModesTarget),
			ModeDoing("You gain 5 life.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return GainLife{Player: item.Controller, Amount: 5}.Apply(ctx)
				}),
		),
	})
}
