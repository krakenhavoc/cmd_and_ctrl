package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "4b072fbb-5e61-4009-bcb1-d493eb4a1be4",
		Name:            "Quantum Entanglement",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventBeginEndStep}, quantumEntanglementWhen,
				quantumEntanglementLabel, quantumEntanglementPayOffer),
		},
	})
}

const quantumEntanglementLabel = "Quantum Entanglement — pay {1}{W} to exile target creature you control, then return it"

// quantumEntanglementWhen is "this enchantment enters" or "your end step
// begins".
func quantumEntanglementWhen(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
	if ev.Kind == game.EventETB {
		return effects.Self(ev, source, ch, g)
	}
	return effects.ByYou(ev, source, ch, g)
}

func quantumEntanglementPayOffer(g *game.Game, item *game.StackItem) error {
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}{W}",
		Question: quantumEntanglementLabel + "?",
		Then:     quantumEntanglementPaid,
	}.Apply(effects.NewContext(g, item))
}

var quantumEntanglementPaid = effects.PayUnlessBranches("pay-unless/quantum-entanglement-blink",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.ReflexiveTrigger{
			Label: "Quantum Entanglement — exile target creature you control, then return that card to the battlefield under its owner's control",
			Body:  quantumEntanglementBlinkBody,
		}.Apply(ctx)
	}, nil)

var quantumEntanglementBlinkBody = game.ReflexiveBody("quantum-entanglement/blink",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				return effects.Flicker{Target: t.ID}.Apply(ctx)
			}
		}
		return nil
	}),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.PermanentYouControl("target creature you control", effects.Creature())
	}))
