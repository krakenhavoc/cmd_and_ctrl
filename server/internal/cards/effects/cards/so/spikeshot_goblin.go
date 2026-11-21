package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spikeshot Goblin — Creature — Goblin Shaman {2}{R}, 1/2 (Commander
// Masters):
//
//	"{R}, {T}: This creature deals damage equal to its power to any target."
//
// Spikeshot Elder's body with a tap in the cost: the power is read as the
// ability resolves, or as the Goblin last existed (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3cbd4c8e-fe3f-4a7c-90a5-7bc656d7bb10",
		Name:         "Spikeshot Goblin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, {T}: This creature deals damage equal to its power to any target.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
			Targets: effects.TargetAny(),
			Effect:  effects.SourceDealsDamageEqualToItsPowerToTarget,
		}},
	})
}
