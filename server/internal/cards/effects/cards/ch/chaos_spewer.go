package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chaos Spewer — Creature — Goblin Warlock {2}{B/R} (Lorwyn Eclipsed, #4458):
//
//	"When this creature enters, you may pay {2}. If you don't, blight 2."
//
// The pay is MayPay with only a decline branch; declining (or being unable
// to pay) blights 2: both -1/-1 counters on one creature you control, which
// may be the Spewer itself (CR 701.68).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77217a66-4ba8-443f-b915-178cfcc233e1",
		Name:         "Chaos Spewer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Chaos Spewer — you may pay {2}; if you don't, blight 2", func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{2}",
					Question: "Chaos Spewer — pay {2}? (If you don't, blight 2.)",
					Then:     chaosSpewerPaid,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

// chaosSpewerPaid has only the decline branch: "if you don't, blight 2". An
// on-disk key: never renamed, never reused.
var chaosSpewerPaid = effects.PayUnlessBranches("pay-unless/chaos-spewer-blight", nil,
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.Blight{N: 2}.Apply(ctx)
	})
