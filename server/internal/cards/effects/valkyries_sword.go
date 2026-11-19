package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valkyrie's Sword — Artifact — Equipment {1}{W} (Foundations Jumpstart):
//
//	"When this Equipment enters, you may pay {4}{W}. If you do, create a
//	 4/4 white Angel Warrior creature token with flying and vigilance, then
//	 attach this Equipment to it.
//	 Equipped creature gets +2/+1.
//	 Equip {3}"
//
// Giant's Amulet's optional payment (MayPay) and keyed paid half, with the
// Angel Warrior. The attach rides the token's creation; if the Sword has
// left in response the token is made and nothing is attached (CR 701.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "236c4ff1-8c3f-4763-8df9-925021e65726",
		Name:         "Valkyrie's Sword",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 1)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Valkyrie's Sword — you may pay {4}{W} to create a 4/4 Angel Warrior and attach this to it",
				valkyriesSwordMayPay),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}

// valkyriesSwordMayPay is the enters trigger's body. Package-level, so it
// captures nothing across an undo.
func valkyriesSwordMayPay(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{4}{W}",
		Question: "Valkyrie's Sword — pay {4}{W} to create a 4/4 white Angel Warrior with flying and vigilance and attach the Sword to it?",
		Then:     valkyriesSwordPaid,
	}.Apply(NewContext(g, item))
}

// valkyriesSwordPaid is the paid half, keyed. An on-disk key.
var valkyriesSwordPaid = PayUnlessBranches("pay-unless/valkyries-sword-angel-and-attach",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return createTokenAndAttachSource(ctx.Game, ctx.Item, "4/4 white Angel Warrior with flying and vigilance", true)
	}, nil)
