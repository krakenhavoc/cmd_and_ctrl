package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Apocalypse Hydra — Creature — Hydra {X}{R}{G}, 0/0 (Battlebond, #5529):
//
//	This creature enters with X +1/+1 counters on it. If X is 5 or more, it enters with an additional X +1/+1 counters on it.
//	{1}{R}, Remove a +1/+1 counter from this creature: It deals 1 damage to any target.
//
// Two entry-counter clauses off the cast X (Walking Ballista's, and one that counts only when X is 5 or more); the ping is Walking Ballista's second ability with a {1}{R} price.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a840210-204d-41cd-afc0-46abf7ea5f87",
		Name:         "Apocalypse Hydra",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			effects.XCounters(game.CounterPlusOne),
			{
				Kind: game.CounterPlusOne,
				Count: func(cast game.CastCounts) int {
					if cast.X >= 5 {
						return cast.X
					}
					return 0
				},
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, Remove a +1/+1 counter from this creature: It deals 1 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.B33DamageChosenTargetFromSource(1),
		}},
	})
}
