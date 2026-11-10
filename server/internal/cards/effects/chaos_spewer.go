package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chaos Spewer — Creature — Goblin Warlock {2}{B/R} (Lorwyn Eclipsed, #4458):
//
//	"When this creature enters, you may pay {2}. If you don't, blight 2."
//
// The pay is MayPay with only a decline branch; declining (or being unable
// to pay) blights 2: both -1/-1 counters on one creature you control, which
// may be the Spewer itself (CR 701.68).
func init() {
	Register(Spec{
		OracleID:     "77217a66-4ba8-443f-b915-178cfcc233e1",
		Name:         "Chaos Spewer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Chaos Spewer — you may pay {2}; if you don't, blight 2", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{2}",
					Question: "Chaos Spewer — pay {2}? (If you don't, blight 2.)",
					Then:     chaosSpewerPaid,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

// chaosSpewerPaid has only the decline branch: "if you don't, blight 2". An
// on-disk key: never renamed, never reused.
var chaosSpewerPaid = PayUnlessBranches("pay-unless/chaos-spewer-blight", nil,
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return Blight{N: 2}.Apply(ctx)
	})
