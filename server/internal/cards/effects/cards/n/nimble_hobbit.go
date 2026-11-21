package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nimble Hobbit — Creature — Halfling Peasant {1}{W}, 1/3 (The Lord of the
// Rings: Tales of Middle-earth Commander):
//
//	"Whenever this creature attacks, you may sacrifice a Food or pay
//	 {2}{W}. When you do, tap target creature an opponent controls."
//
// Anthropede's may-pay with an either payment: one prompt whose answer names
// the branch, with a Food to sacrifice or {2}{W} to pay, and a branch that
// can't be paid not offered (CR 118.3). Paying creates a CR 603.12
// reflexive trigger whose target is chosen as it goes on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "547da0ef-f1e1-4b04-ba83-4845552b0dc4",
		Name:         "Nimble Hobbit",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Nimble Hobbit — you may sacrifice a Food or pay {2}{W}", func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Pay:      effects.GapB128SacrificeAFoodOrPay("{2}{W}"),
					Question: "Nimble Hobbit — sacrifice a Food or pay {2}{W} to tap target creature an opponent controls?",
					Then:     nimbleHobbitPaid,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

var nimbleHobbitPaid = effects.PayUnlessBranches("pay-unless/nimble-hobbit-tap",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.WhenYouDo("Nimble Hobbit — tap target creature an opponent controls", nimbleHobbitTapBody).Apply(ctx)
	}, nil)

var nimbleHobbitTapBody = game.ReflexiveBody("nimble-hobbit/tap", effects.SimpleBody(effects.B36TapChosenCreature),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())
	}))
