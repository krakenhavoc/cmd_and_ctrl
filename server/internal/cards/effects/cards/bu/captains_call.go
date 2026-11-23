package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Captain's Call — Sorcery {3}{W}:
//
//	"Create three 1/1 white Soldier creature tokens."
//
// Raise the Alarm's token clause at sorcery speed, three of them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "46418fe4-065c-4dfa-b796-eee02c14f351",
		Name:         "Captain's Call",
		Purpose:      game.Purpose{Tokens: 3},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{
				Controller: ctx.Controller(),
				Template:   effects.TokenCard("1/1 white Soldier"),
				N:          3,
			}.Apply(ctx)
		},
	})
}
