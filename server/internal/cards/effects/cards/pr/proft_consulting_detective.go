package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Proft, Consulting Detective — Legendary Creature — Human Detective
// {1}{U}, 2/2:
//
//	"Whenever you scry or surveil, you may pay {2}. If you do, put a
//	 +1/+1 counter on Proft and draw a card. (Draw after you scry or
//	 surveil.)"
//
// One ability watching EventScry and EventSurveil, each emitted once the
// scry or surveil is complete, so the card is drawn after it, as the
// reminder text says. The {2} is asked as the trigger resolves
// (MayPay). The counter goes on Proft only while he is still the same
// object on the battlefield; the draw happens regardless.
//
// No simplification.
// proftConsultingDetectivePaid is the counter and then the draw: the draw
// waits for a prompt the counter asks (a CR 616 order; CR 608.2c, #4155).
var proftConsultingDetectivePaid = effects.RegisterThenSteps("then-steps/proft-consulting-detective-paid",
	effects.Instruction(func(ctx *effects.Context) error {
		if !effects.B09SourceStillOnBattlefield(ctx.Game, ctx.Item) {
			return nil
		}
		return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}),
	effects.DrawCards{N: 1},
)

func init() {
	const label = "Proft, Consulting Detective — you may pay {2} for a +1/+1 counter and a card"
	effects.Register(effects.Spec{
		OracleID:     "45473bdb-b96f-4213-a46d-b2dbdf6b5e56",
		Name:         "Proft, Consulting Detective",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventScry, game.EventSurveil}, effects.ByYou, label,
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MayPay{
						Chooser:  ctx.Controller(),
						Cost:     "{2}",
						Question: "Proft, Consulting Detective — pay {2} to put a +1/+1 counter on Proft and draw a card?",
						Then:     proftPaid,
					}.Apply(ctx)
				}),
		},
	})
}

// proftPaid is the paid half, keyed (#4264): the counter and the draw are the keyed steps above, run against the asking item.
var proftPaid = effects.PayUnlessBranches("pay-unless/proft-consulting-detective-counter-and-draw",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error { return proftConsultingDetectivePaid.Run(ctx) }, nil)
