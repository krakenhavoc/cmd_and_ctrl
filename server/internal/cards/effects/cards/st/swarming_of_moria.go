package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Swarming of Moria — Sorcery {2}{R} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Create a Treasure token.
//	 Amass Orcs 2."
//
// The Treasure, then the keyword action (CR 701.47), as steps of one
// resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c96402dc-1509-4136-bbfd-c86a85e3ee6f",
		Name:         "Swarming of Moria",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Amass: 2, Tokens: 1, Answers: game.AnswerValue},
		OnResolve: effects.ResolveSteps(
			effects.CreateToken{Template: effects.TreasureToken(), N: 1},
			effects.Amass{Subtype: "Orc", N: 2},
		),
	})
}
