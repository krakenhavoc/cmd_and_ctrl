package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "0f1ada92-538b-40cb-9bf3-5c415f5325e6",
		Name:         "Spotcycle Scouter",
		Purpose:      game.Purpose{Scry: 2},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Spotcycle Scouter — scry 2", effects.Do(effects.Scry{N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Spotcycle Scouter"),
		}},
	})
}
