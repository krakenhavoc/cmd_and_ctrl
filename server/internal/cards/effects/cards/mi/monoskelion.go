package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Monoskelion — Artifact Creature — Construct {2}, 1/1 (Modern Horizons 2,
// #5501):
//
//	"This creature enters with a +1/+1 counter on it.
//	 {1}, Remove a +1/+1 counter from this creature: It deals 1 damage to any
//	 target."
//
// The entry counter is a self-replacement (CR 614.1c) and the counter is
// removed as a cost (Walking Ballista's), so a response cannot spend it twice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c16d79d9-c159-421f-90a2-13313a1de5f6",
		Name:         "Monoskelion",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.OldB031SelfEntersWithCounters(game.CounterPlusOne, 1)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Remove a +1/+1 counter from this creature: It deals 1 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Targets: effects.TargetAny(),
			Effect:  effects.B33DamageChosenTargetFromSource(1),
		}},
	})
}
