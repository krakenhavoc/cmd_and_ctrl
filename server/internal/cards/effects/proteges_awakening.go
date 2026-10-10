package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Protege's Awakening — Sorcery {3}{U} (Reality Fracture):
//
//	"Empower Jace 6.
//	 Draw a card."
//
// ADR 0139: the draw is printed after the keyword action, so it is the
// action's Then and waits for the Jace choice and any counter
// replacement ordering.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "07617648-6490-4719-acec-7705671f7e01",
		Name:         "Protege's Awakening",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return EmpowerJace{N: 6, Then: func(ctx *Context) error {
				return DrawCards{N: 1}.Apply(ctx)
			}}.Apply(ctx)
		},
	})
}
