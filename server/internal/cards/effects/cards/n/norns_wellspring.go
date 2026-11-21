package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Norn's Wellspring — Artifact {1}{W}:
//
//	"Whenever a creature you control dies, scry 1 and put an oil counter
//	 on this artifact.
//	 {1}, {T}, Remove two oil counters from this artifact: Draw a card."
//
// One trigger doing both halves: the scry is a prompt, so the counter waits
// for its answer (CR 608.2c). The counter lands only while the Wellspring is
// still the permanent that triggered. The activation pays {1}, the tap and
// exactly two oil counters up front, so one with fewer cannot be activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56a55f45-2cc6-407d-b212-5619c98456d9",
		Name:         "Norn's Wellspring",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.WheneverACreatureYouControlDies("Norn's Wellspring — scry 1 and put an oil counter on it",
					effects.Do(
						effects.Instruction(func(ctx *effects.Context) error {
							return effects.Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
						}),
						effects.CounterOnThis{Kind: effects.S82OneB12Oil, N: 1},
					)),
				game.Purpose{DeathPayoff: true}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Remove two oil counters from this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.RemoveCountersFromThis(effects.S82OneB12Oil, 2)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.DrawNForEffect(item.Controller, 1)
			},
		}},
	})
}
