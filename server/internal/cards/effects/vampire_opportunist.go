package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vampire Opportunist — Creature — Vampire {1}{B}, 2/1 (War of the Spark, #5520):
//
//	"{6}{B}: Each opponent loses 2 life and you gain 2 life."
//
// Bloodsoaked Reveler's drain, at {6}{B}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f525ccd-4776-492c-a1f9-8bce5740fb28",
		Name:         "Vampire Opportunist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{6}{B}: Each opponent loses 2 life and you gain 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue, LifeGain: 2},
			Cost:    ManaCost("{6}{B}"),
			Effect:  Do(EachOpponentLosesLife{N: 2}, GainLife{Amount: 2}),
		}},
	})
}
