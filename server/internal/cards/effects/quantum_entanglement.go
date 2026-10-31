package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quantum Entanglement — {1}{W} Enchantment (Marvel Super Heroes, #4455):
//
//	"Flash
//	 When this enchantment enters and at the beginning of your end step, you
//	 may pay {1}{W}. When you do, exile target creature you control, then
//	 return that card to the battlefield under its owner's control."
//
// One ability on two events (Quantum Entanglement's enters + your end step),
// so one registered trigger. The "may pay" is Matt Murdock's MayPay with a
// ReflexiveTrigger (CR 603.12): the target is chosen only when the payment
// was made, and a creature that has gone by resolution is left alone
// (CR 608.2b). The blink is Flicker, which returns the card under its
// owner's control as a new object (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4b072fbb-5e61-4009-bcb1-d493eb4a1be4",
		Name:            "Quantum Entanglement",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventBeginEndStep}, quantumEntanglementWhen,
				quantumEntanglementLabel, quantumEntanglementPayOffer),
		},
	})
}

const quantumEntanglementLabel = "Quantum Entanglement — pay {1}{W} to exile target creature you control, then return it"

// quantumEntanglementWhen is "this enchantment enters" or "your end step
// begins".
func quantumEntanglementWhen(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
	if ev.Kind == game.EventETB {
		return Self(ev, source, ch, g)
	}
	return ByYou(ev, source, ch, g)
}

func quantumEntanglementPayOffer(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}{W}",
		Question: quantumEntanglementLabel + "?",
		Then:     quantumEntanglementPaid,
	}.Apply(NewContext(g, item))
}

var quantumEntanglementPaid = PayUnlessBranches("pay-unless/quantum-entanglement-blink",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return ReflexiveTrigger{
			Label: "Quantum Entanglement — exile target creature you control, then return that card to the battlefield under its owner's control",
			Body:  quantumEntanglementBlinkBody,
		}.Apply(ctx)
	}, nil)

var quantumEntanglementBlinkBody = game.ReflexiveBody("quantum-entanglement/blink",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				return Flicker{Target: t.ID}.Apply(ctx)
			}
		}
		return nil
	}),
	constTargets(func() *game.TargetSpec { return PermanentYouControl("target creature you control", Creature()) }))
