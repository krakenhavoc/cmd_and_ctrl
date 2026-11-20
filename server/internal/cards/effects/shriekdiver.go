package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shriekdiver — Creature — Zombie Bird Warrior {2}{B}, 2/1 (War of the Spark, #5520):
//
//	"Flying
//	 {1}: This creature gains haste until end of turn."
//
// Flying is printed data; the haste is a layer-6 grant on itself until
// end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f51e1660-e1c8-4122-8c43-321c820a1fc1",
		Name:            "Shriekdiver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{1}: This creature gains haste until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{1}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("haste", "Shriekdiver — haste until end of turn"),
		}},
	})
}
