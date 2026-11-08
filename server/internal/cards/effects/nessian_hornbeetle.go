package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nessian Hornbeetle — Creature — Insect {1}{G}, 2/2 (Foundations, #4465):
//
//	"At the beginning of combat on your turn, if you control another
//	 creature with power 4 or greater, put a +1/+1 counter on this
//	 creature."
//
// Nasty Little Rabbit's ferocious trigger with "another": the Hornbeetle's
// own power never counts, however big it has grown. An intervening "if",
// checked as combat begins and again as the trigger resolves, against
// powers with every layer applied.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ab097f46-a7bf-4cf1-a96c-36b681fe79da",
		Name:         "Nessian Hornbeetle",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control another creature with power 4 or greater",
				Test: func(ctx *Context) bool {
					return hornbeetleControlsAnotherPowerFour(ctx.Game, ctx.Controller(), ctx.Source())
				},
			}, On(game.EventStepBegan, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return StepBegan(game.StepBeginCombat, true)(ev, source, lki, g) &&
					hornbeetleControlsAnotherPowerFour(g, source.Controller, source.InstanceID)
			}, "Nessian Hornbeetle — a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}

// hornbeetleControlsAnotherPowerFour is "you control another creature
// with power 4 or greater": a creature `controller` controls, other than
// `self`, with layered power 4 or more.
func hornbeetleControlsAnotherPowerFour(g *game.Game, controller, self uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != self && c.Controller == controller && c.IsCreature() && c.CurrentPower() >= 4 {
			return true
		}
	}
	return false
}
