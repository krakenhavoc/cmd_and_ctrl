package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "1cf4e16b-2273-41c7-9bcd-0cdc78e12198",
		Name:         "Unassuming Sage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Unassuming Sage — you may pay {2} to create a Sorcerer Role token attached to it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}",
						Question: "Unassuming Sage — pay {2} to create a Sorcerer Role token attached to it?",
						Then:     unassumingSagePaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// unassumingSagePaid is the paid half, keyed (#4264).
var unassumingSagePaid = effects.PayUnlessBranches("pay-unless/unassuming-sage-sorcerer-role",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateRoleOnThis(effects.RoleSorcerer)(ctx.Game, ctx.Item)
	}, nil)
