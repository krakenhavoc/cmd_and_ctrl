package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stream Hopper — Creature — Goblin {U/R}, 1/1 (Eventide, #5590):
//
//	"{U/R}: This creature gains flying until end of turn."
//
// Killer Whale's self grant at a hybrid cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "153cd93c-7bba-4482-893c-55a602c4439b",
		Name:         "Stream Hopper",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U/R}: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{U/R}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("flying", "Stream Hopper — flying until end of turn"),
		}},
	})
}
