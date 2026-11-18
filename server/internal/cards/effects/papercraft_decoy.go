package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Papercraft Decoy — Artifact Creature — Frog {2}, 2/1 (Kamigawa: Neon
// Dynasty, S82):
//
//	"When this creature leaves the battlefield, you may pay {2}. If you do,
//	 draw a card."
//
// Krovod Haunch's may-pay on a leave trigger (WhenThisLeaves: any exit, not
// only dying). The payment is made as the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0494c34-f142-441f-b7f8-e981416f0bad",
		Name:         "Papercraft Decoy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisLeaves("Papercraft Decoy — you may pay {2} to draw a card", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{2}",
					Question: "Papercraft Decoy — pay {2} to draw a card?",
					Then:     papercraftDecoyDraw,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

var papercraftDecoyDraw = PayUnlessBranches("pay-unless/papercraft-decoy-draw",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
	}, nil)
