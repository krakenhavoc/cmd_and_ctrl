package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thatcher Revolt — {2}{R} Sorcery: "Create three 1/1 red Human creature
// tokens with haste. Sacrifice those tokens at the beginning of the next end
// step."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "83cdff93-531d-4148-bff7-5312b3b53bf0",
		Name:         "Thatcher Revolt",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3, Answers: game.AnswerMakesBlocker},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ThatcherRevoltTokens(ctx.Game, item)
		},
	})
}
