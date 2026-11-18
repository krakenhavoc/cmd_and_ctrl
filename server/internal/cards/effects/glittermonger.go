package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glittermonger (Streets of New Capenna, #5491):
//
//	{T}: Create a Treasure token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "184c763c-b52a-471c-bd87-d01729ebe5e4",
		Name:         "Glittermonger",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Cost:    TapCost(),
			Effect:  Do(CreateToken{Template: TreasureToken(), N: 1}),
		}},
	})
}
