package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spellbook Vendor — Creature — Human Peasant {1}{W}, 2/2:
//
//	"Vigilance
//	 At the beginning of combat on your turn, you may pay {1}. When you
//	 do, create a Sorcerer Role token attached to target creature you
//	 control. (If you control another Role on it, put that one into the
//	 graveyard. Enchanted creature gets +1/+1 and has "Whenever this
//	 creature attacks, scry 1.")"
//
// "When you do" is a CR 603.12 reflexive trigger created once the
// payment is made, so the target is chosen as it goes on the stack and
// nothing is targeted when the player declines.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "51dad7d3-f91d-4bd8-aaed-235da42a448b",
		Name:            "Spellbook Vendor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Spellbook Vendor — you may pay {1}", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{1}",
					Question: "Spellbook Vendor — pay {1} to create a Sorcerer Role token attached to a creature you control?",
					OnPay: func(ctx *Context) error {
						return WhenYouDo("Spellbook Vendor — create a Sorcerer Role token attached to target creature you control",
							spellbookVendorSorcererBody).Apply(ctx)
					},
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
