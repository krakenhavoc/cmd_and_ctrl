package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Release the Dogs — Sorcery {3}{W} (Foundations, #4465):
//
//	"Create four 1/1 white Dog creature tokens."
//
// The Dog row is the shared "1/1 white Dog" token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b97d21b4-d730-479f-872f-3e1645b66751",
		Name:         "Release the Dogs",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 4},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 white Dog"), N: 4}.Apply(ctx)
		},
	})
}
