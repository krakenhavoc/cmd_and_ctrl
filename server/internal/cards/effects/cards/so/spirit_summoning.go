package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spirit Summoning — Sorcery — Lesson {1}{R/W}{R/W} (Strixhaven: School of
// Mages, #5502):
//
//	"Create a 3/2 red and white Spirit creature token."
//
// Quintorius Kand's Spirit from the shared token table. Being a Lesson is
// its type line, which learn reads.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a4825093-ea29-47d9-b52c-1c7735f1909e",
		Name:         "Spirit Summoning",
		Purpose:      game.Purpose{Tokens: 1},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("3/2 red and white Spirit"), N: 1}.Apply(ctx)
		},
	})
}
