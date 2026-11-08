package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spotcycle Scouter — Artifact — Vehicle {1}{W}, 3/2 (Aetherdrift,
// #4464):
//
//	"When this Vehicle enters, scry 2.
//	 Crew 1"
//
// Veteran Motorist's enters scry on a Vehicle, and the engine's crew.
// The enters trigger does not ask that the Vehicle be a creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0f1ada92-538b-40cb-9bf3-5c415f5325e6",
		Name:         "Spotcycle Scouter",
		Purpose:      game.Purpose{Scry: 2},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Spotcycle Scouter — scry 2", Do(Scry{N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Spotcycle Scouter"),
		}},
	})
}
