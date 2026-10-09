package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unassuming Sage — Creature — Human Peasant Wizard {1}{W}, 2/2:
//
//	"When this creature enters, you may pay {2}. If you do, create a
//	 Sorcerer Role token attached to it. (Enchanted creature gets +1/+1
//	 and has "Whenever this creature attacks, scry 1.")"
//
// The payment is the MayPay prompt, so it is made as the trigger
// resolves, and the Role is attached to the Sage only if it is still
// a creature on the battlefield by then.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "1cf4e16b-2273-41c7-9bcd-0cdc78e12198",
		Name:         "Unassuming Sage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Unassuming Sage — you may pay {2} to create a Sorcerer Role token attached to it",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}",
						Question: "Unassuming Sage — pay {2} to create a Sorcerer Role token attached to it?",
						OnPay: func(ctx *Context) error {
							return createRoleOnThis(RoleSorcerer)(ctx.Game, ctx.Item)
						},
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
