package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whiptongue Frog — Creature — Frog {2}{U}, 1/3 (Exodus, #5626):
//
//	"{U}: This creature gains flying until end of turn."
//
// Killer Whale's row on a smaller body: the shared
// thisGainsKeywordUntilEndOfTurn grant, ending in the cleanup step
// (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "edc961bd-5ec9-47ed-9d34-e26a48c13702",
		Name:         "Whiptongue Frog",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U}: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{U}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("flying", "Whiptongue Frog — flying"),
		}},
	})
}
