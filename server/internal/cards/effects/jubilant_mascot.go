package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jubilant Mascot — Creature — Homunculus {2}{W}, 1/1 (Foundations
// Jumpstart, BNS batch 32):
//
//	"At the beginning of combat on your turn, you may pay {3}{W}. If you
//	 do, support 2. (Put a +1/+1 counter on each of up to two other target
//	 creatures.)"
//
// Support is a permanent's here, so the targets are other creatures
// (CR 701.41a), up to two, chosen as the trigger goes on the stack
// (CR 603.3d) and so before the payment is asked for. The {3}{W} is paid
// as the trigger resolves (CR 118.12); only a paid payment puts a counter
// on each target still legal (SupportEffect, CR 608.2b). Declining, or
// being unable to pay, puts none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d160887f-f78e-4a66-9766-0b7dd852a784",
		Name:         "Jubilant Mascot",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(AtBeginningOfYourCombat(jubilantMascotLabel, jubilantMascotOffer), SupportTargets(2, true)),
		},
	})
}

const jubilantMascotLabel = "Jubilant Mascot — pay {3}{W} to support 2"

func jubilantMascotOffer(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{3}{W}",
		Question: jubilantMascotLabel + "?",
		Then:     jubilantMascotPaid,
	}.Apply(NewContext(g, item))
}

// jubilantMascotPaid is the paid half, keyed: support runs against the asking
// trigger's own targets.
var jubilantMascotPaid = PayUnlessBranches("pay-unless/jubilant-mascot-support",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		if ctx.Item == nil {
			return nil
		}
		return SupportEffect(ctx.Game, ctx.Item)
	}, nil)
