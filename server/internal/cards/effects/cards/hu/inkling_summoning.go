package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inkling Summoning — Sorcery — Lesson {1}{W/B}{W/B} (Strixhaven: School
// of Mages, #5502):
//
//	"Create a 2/1 white and black Inkling creature token with flying."
//
// Intermediate Chirography's Inkling from the shared token table. Being a
// Lesson is its type line, which learn reads.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "999a9310-d145-4c1a-8518-9d96602ca672",
		Name:         "Inkling Summoning",
		Purpose:      game.Purpose{Tokens: 1},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("2/1 white and black Inkling with flying"), N: 1}.Apply(ctx)
		},
	})
}
