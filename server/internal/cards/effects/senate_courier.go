package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Senate Courier — Creature — Bird {2}{U}, 1/4 (Ravnica Allegiance, #5521):
//
//	"Flying
//	 {1}{W}: This creature gains vigilance until end of turn."
//
// Flying is printed data (CR 702.9b); the vigilance is a layer-6 grant
// until end of turn on itself (thisGainsKeywordUntilEndOfTurn).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4429e14f-a5f7-42e9-a8c3-b59428f92047",
		Name:            "Senate Courier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: This creature gains vigilance until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{1}{W}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("vigilance", "Senate Courier — vigilance until end of turn"),
		}},
	})
}
