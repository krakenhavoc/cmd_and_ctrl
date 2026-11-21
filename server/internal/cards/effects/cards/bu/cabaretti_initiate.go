package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cabaretti Initiate (Streets of New Capenna, #5491):
//
//	{2}{R/W}: This creature gains double strike until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e0d877c-6ce3-462b-9746-7fc921c93be9",
		Name:         "Cabaretti Initiate",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R/W}: This creature gains double strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{2}{R/W}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("double strike", "Cabaretti Initiate — double strike until end of turn"),
		}},
	})
}
