package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Welder Automaton — Artifact Creature — Construct {2}, 2/1 (Aether Revolt,
// #5536):
//
//	"{3}{R}: This creature deals 1 damage to each opponent."
//
// Fire Nation Archers' DamageEachOpponent without the token: the
// Automaton is the source, one damage event per opponent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "52eae443-fbe5-479f-a5a4-b5e4401d62d3",
		Name:         "Welder Automaton",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{R}: This creature deals 1 damage to each opponent.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{3}{R}"),
			Effect:  Do(DamageEachOpponent{N: 1}),
		}},
	})
}
