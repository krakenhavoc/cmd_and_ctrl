package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spined Tyrranax — Creature — Dinosaur Beast {4}{G}, 5/5 (Foundations
// Jumpstart):
//
//	"At the beginning of combat on your turn, you may pay {2}{G}. When you
//	 do, put a +1/+1 counter on target creature. That creature gains
//	 trample until end of turn."
//
// Matt Murdock's shape: a may-pay at the beginning of your combat, a
// reflexive trigger on payment that targets any creature (CR 603.12).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2d09e689-c163-48c5-af5d-766f110d041e",
		Name:         "Spined Tyrranax",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat(spinedTyrranaxLabel, spinedTyrranaxOffer),
		},
	})
}

const spinedTyrranaxLabel = "Spined Tyrranax — you may pay {2}{G}; when you do, +1/+1 counter and trample on target creature"

func spinedTyrranaxOffer(g *game.Game, item *game.StackItem) error {
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{2}{G}",
		Question: spinedTyrranaxLabel + "?",
		Then:     spinedTyrranaxPaid,
	}.Apply(effects.NewContext(g, item))
}

var spinedTyrranaxPaid = effects.PayUnlessBranches("pay-unless/spined-tyrranax-counter",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.ReflexiveTrigger{
			Label: "Spined Tyrranax — +1/+1 counter and trample on target creature",
			Body:  spinedTyrranaxCounterBody,
		}.Apply(ctx)
	}, nil)

var spinedTyrranaxCounterBody = game.ReflexiveBody("spined-tyrranax/counter", effects.SimpleBody(effects.SpinedTyrranaxCounterAndTrample),
	effects.ConstTargets(func() *game.TargetSpec { return effects.TargetCreature("target creature") }))
