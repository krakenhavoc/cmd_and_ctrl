package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wretched Doll — Artifact Creature — Toy {1}{B}, 3/1 (Aetherdrift,
// #4464):
//
//	"{B}, {T}: Surveil 1. (Look at the top card of your library. You may
//	 put that card into your graveyard.)"
//
// An ordinary activated ability that uses the stack: surveil 1
// (CR 701.25a). The {T} means summoning sickness stops it (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "628cf9ab-8f8b-49b6-b761-2450088ad03b",
		Name:         "Wretched Doll",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}, {T}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{B}"), TapCost()),
			Effect:  Do(Surveil{N: 1}),
		}},
	})
}
