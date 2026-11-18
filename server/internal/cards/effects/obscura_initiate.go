package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Obscura Initiate (Streets of New Capenna, #5491):
//
//	Flying
//	{1}{W/B}: This creature gains lifelink until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bdc3a563-86e0-4457-b00b-85c03e62a930",
		Name:            "Obscura Initiate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W/B}: This creature gains lifelink until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{1}{W/B}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("lifelink", "Obscura Initiate — lifelink until end of turn"),
		}},
	})
}
