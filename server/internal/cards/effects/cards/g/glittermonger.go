package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glittermonger (Streets of New Capenna, #5491):
//
//	{T}: Create a Treasure token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "184c763c-b52a-471c-bd87-d01729ebe5e4",
		Name:         "Glittermonger",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Cost:    effects.TapCost(),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1}),
		}},
	})
}
