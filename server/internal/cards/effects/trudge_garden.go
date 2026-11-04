package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trudge Garden — Enchantment {2}{G} (Secrets of Strixhaven Commander,
// #4834):
//
//	"Whenever you gain life, you may pay {2}. If you do, create a 4/4 green
//	 Fungus Beast creature token with trample."
//
// Dawn of Hope's shape: the trigger fires once per life-gain event (lifelink
// included, so three connecting lifelinkers are three prompts), and the {2}
// is paid on resolution (CR 603.5), with the token made only when the mana was
// really paid. The Fungus Beast is the 4/4 green token printed for the set.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "be175e92-567a-42e3-ad54-136c0631fe0e",
		Name:         "Trudge Garden",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Trudge Garden — pay {2} to create a 4/4 green Fungus Beast token with trample", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{2}",
					Question: "Trudge Garden — pay {2} to create a 4/4 green Fungus Beast creature token with trample?",
					Then:     socB14TrudgeGardenPaid,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

// socB14TrudgeGardenPaid is the paid half, keyed: the token comes to the
// controller of the asking item.
var socB14TrudgeGardenPaid = PayUnlessBranches("pay-unless/trudge-garden-fungus-beast",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return CreateToken{Controller: ctx.Controller(), Template: TokenCard("4/4 green Fungus Beast with trample"), N: 1}.Apply(ctx)
	}, nil)
