package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tolarian Kraken — Creature — Kraken {4}{U}{U}, 4/6 (Jumpstart):
//
//	"Whenever you draw a card, you may pay {1}. When you do, you may tap
//	 or untap target creature."
//
// Nadir Kraken's draw trigger and payment, one trigger per card drawn,
// with the "when you do" as a CR 603.12 reflexive trigger created only on
// the branch where {1} was paid. Its target is chosen as it goes on the
// stack, above the draw trigger, so the table gets a window between the
// two. "You may tap or untap" is Component Collector's reading: the pick
// between the two is made as the reflexive trigger resolves, and tapping a
// tapped creature or untapping an untapped one does nothing, so the "may"
// needs no third option. A target gone by then takes the clause with it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "846887e4-744f-44ce-9110-d09680b11e95",
		Name:         "Tolarian Kraken",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouDraw("Tolarian Kraken — you may pay {1}; when you do, you may tap or untap target creature",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Tolarian Kraken — pay {1} to tap or untap target creature?",
						Then:     tolarianKrakenPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// tolarianKrakenPaid is the paid half, keyed (#4264): the reflexive
// trigger.
var tolarianKrakenPaid = PayUnlessBranches("pay-unless/tolarian-kraken-tap-or-untap",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return WhenYouDo("Tolarian Kraken — you may tap or untap target creature", tolarianKrakenTapBody).Apply(ctx)
	}, nil)

// tolarianKrakenTapBody is the reflexive trigger's body and clause.
var tolarianKrakenTapBody = game.ReflexiveBody("tolarian-kraken/tap-or-untap",
	simpleBody(b38TapOrUntapTarget("Tolarian Kraken")),
	constTargets(func() *game.TargetSpec { return TargetCreature("target creature") }))
