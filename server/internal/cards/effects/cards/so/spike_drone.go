package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spike Drone — Creature — Spike Drone {G}, 0/0 (Tempest, S82 OLD batch
// 760, #5628):
//
//	"This creature enters with a +1/+1 counter on it.
//	 {2}, Remove a +1/+1 counter from this creature: Put a +1/+1 counter on
//	 target creature."
//
// Benevolent Hydra's counter-removal cost and Renew's counter on a target.
// The counter is removed at announcement (CR 602.2b), so the Drone, a 0/0
// without it, may die before the ability resolves; the counter still goes on
// the target. It can target itself (it then has the counter back, if it
// survived).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd6d15c0-c294-42c9-a6cf-57b0b5d40823",
		Name:         "Spike Drone",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 1, "Spike Drone: enters with a +1/+1 counter"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Remove a +1/+1 counter from this creature: Put a +1/+1 counter on target creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.RenewCountersOnTarget(game.CounterPlusOne, 1),
		}},
	})
}
