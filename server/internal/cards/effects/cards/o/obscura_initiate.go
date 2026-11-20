package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Obscura Initiate (Streets of New Capenna, #5491):
//
//	Flying
//	{1}{W/B}: This creature gains lifelink until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bdc3a563-86e0-4457-b00b-85c03e62a930",
		Name:            "Obscura Initiate",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W/B}: This creature gains lifelink until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}{W/B}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("lifelink", "Obscura Initiate — lifelink until end of turn"),
		}},
	})
}
