package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "3cbd4c8e-fe3f-4a7c-90a5-7bc656d7bb10",
		Name:         "Spikeshot Goblin",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}, {T}: This creature deals damage equal to its power to any target.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    Plus(ManaCost("{R}"), TapCost()),
			Targets: TargetAny(),
			Effect:  sourceDealsDamageEqualToItsPowerToTarget,
		}},
	})
}
