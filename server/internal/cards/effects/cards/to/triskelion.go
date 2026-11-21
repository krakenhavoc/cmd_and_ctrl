package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Triskelion — Artifact Creature — Construct {6}, 1/1:
//
//	"This creature enters with three +1/+1 counters on it.
//	 Remove a +1/+1 counter from this creature: It deals 1 damage to any
//	 target."
//
// The entry counters are the printed CR 614.1c clause (the 1/1 body
// never exists without them); the ability's cost removes a counter from
// this creature and its target is any target. Same shape as Deathbringer
// Thoctar's ping.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "74f67dcf-5afb-45aa-8d4b-3cdb23f6f2a1",
		Name:         "Triskelion",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 3, "Triskelion: enters with three +1/+1 counters"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove a +1/+1 counter from this creature: It deals 1 damage to any target.",
			Cost:    effects.RemoveCountersFromThis(game.CounterPlusOne, 1),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
