package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elemental Summoning — Sorcery — Lesson {3}{U/R}{U/R} (Strixhaven: School
// of Mages, #5502):
//
//	"Create a 4/4 blue and red Elemental creature token."
//
// Magma Opus's Elemental from the shared token table. Being a Lesson is
// its type line, which learn reads.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f9e668f9-21af-4fd8-9a37-6505ab2234e3",
		Name:         "Elemental Summoning",
		Purpose:      game.Purpose{Tokens: 1},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("4/4 blue and red Elemental"), N: 1}.Apply(ctx)
		},
	})
}
