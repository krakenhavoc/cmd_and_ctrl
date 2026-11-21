package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Icatian Town — Sorcery {5}{W} (Fallen Empires, #5637):
//
//	"Create four 1/1 white Citizen creature tokens."
//
// Release the Dogs with Citizens: the "1/1 white Citizen" row in the
// token table, four of them under the caster's control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa8b60b6-cf55-4aaa-9caa-2b17942d8269",
		Name:         "Icatian Town",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 4},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 white Citizen"), N: 4}.Apply(ctx)
		},
	})
}
