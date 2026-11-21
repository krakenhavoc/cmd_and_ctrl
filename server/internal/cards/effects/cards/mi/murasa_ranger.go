package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Murasa Ranger — Creature — Human Warrior Ranger {3}{G}, 3/3
// (Foundations Jumpstart):
//
//	"Landfall — Whenever a land you control enters, you may pay {3}{G}.
//	 If you do, put two +1/+1 counters on this creature."
//
// The payment is offered as the trigger resolves (Cradle of Vitality's
// MayPay); the counters go on only if the mana was really paid, and
// AddCounter does nothing once the Ranger has left the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84ba3598-a2d2-46d4-b73c-ab139cac37c4",
		Name:         "Murasa Ranger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Murasa Ranger — you may pay {3}{G} to put two +1/+1 counters on it (landfall)",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{3}{G}",
						Question: "Murasa Ranger — pay {3}{G} to put two +1/+1 counters on it?",
						Then:     murasaRangerPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// murasaRangerPaid is the paid half, keyed.
var murasaRangerPaid = effects.PayUnlessBranches("pay-unless/murasa-ranger-counters",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
	}, nil)
