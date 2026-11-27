package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Molten Hydra — Creature — Hydra {1}{R} 1/1:
//
//	"{1}{R}{R}: Put a +1/+1 counter on this creature.
//	 {T}, Remove all +1/+1 counters from this creature: It deals damage to any
//	 target equal to the number of +1/+1 counters removed this way."
//
// Removing the counters is the cost, so the amount is what the announcement
// removed, read back as the ability resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd5bc95d-4915-4c5a-9b57-2730adf69797",
		Name:         "Molten Hydra",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}{R}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{1}{R}{R}"),
			Effect:  effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}, {
			Label:   "{T}, Remove all +1/+1 counters from this creature: It deals damage to any target equal to the number of +1/+1 counters removed this way.",
			Cost:    effects.Plus(effects.TapCost(), effects.RemoveAllCountersFromThis(game.CounterPlusOne)),
			Targets: effects.TargetAny(),
			Effect:  effects.SourceDealsComputedDamageToEachLegalTarget(func(ctx *effects.Context) int { return ctx.CountersRemoved() }),
		}},
	})
}
