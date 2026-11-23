package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quarrel's End — Sorcery {2}{R} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"As an additional cost to cast this spell, discard a card.
//	 Draw two cards and create a 1/1 white Human Soldier creature token."
//
// Big Score's additional discard cost (paid on announcement, so a hand with no
// other card refuses the cast), then the two draws and the token as steps of one
// resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "fc46649d-9d84-4082-9d2d-fbb7ce91c6a4",
		Name:           "Quarrel's End",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Draws: 2, Tokens: 1, Answers: game.AnswerValue},
		AdditionalCost: effects.DiscardCost(1),
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 2},
			effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 1},
		),
	})
}
