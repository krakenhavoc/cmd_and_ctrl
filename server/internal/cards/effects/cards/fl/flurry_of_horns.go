package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flurry of Horns — Sorcery {4}{R} (Jumpstart, #5510):
//
//	"Create two 2/3 red Minotaur creature tokens with haste."
//
// The printed Minotaur token has no abilities: the haste is the spell's, so
// it is granted on top of the printed template.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c8ab7bf-7b6a-4a34-b988-9f3f4d84a9ea",
		Name:         "Flurry of Horns",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2, Answers: game.AnswerMakesBlocker},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateTokenAdvanced{Spec: effects.Token(effects.TokenCard("2/3 red Minotaur")).WithKeywords("haste"), N: 2}.Apply(ctx)
		},
	})
}
