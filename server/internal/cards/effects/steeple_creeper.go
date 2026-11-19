package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Steeple Creeper — Creature — Frog Snake {2}{G}, 4/2 (Ravnica: Clue Edition, #5929): "{3}{U}: This creature gains flying until end of turn."
//
// Killer Whale's shape with a {3}{U} cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f53f33ed-5c8c-4835-bc5f-65a8acfa3f5b",
		Name:         "Steeple Creeper",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{3}{U}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("flying", "Steeple Creeper — flying"),
		}},
	})
}
