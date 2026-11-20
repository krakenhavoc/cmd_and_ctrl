package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tyrranax — Creature — Dinosaur Beast {4}{G}{G}, 5/4 (Fifth Dawn, #5604):
//
//	"{1}{G}: This creature gets -1/+1 until end of turn."
//
// Endling's -1/+1 half at {1}{G}, with the shared thisGetsUntilEndOfTurn
// body: a layer 7c boost on itself that ends at cleanup (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "28ecc209-d78b-4b37-8411-fb8079f6e088",
		Name:         "Tyrranax",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}: This creature gets -1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{G}"),
			Effect:  thisGetsUntilEndOfTurn(-1, 1, "Tyrranax — -1/+1 until end of turn"),
		}},
	})
}
