package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Index — Sorcery {U} (Magic 2013, S82 OLD, #5568):
//
//	"Look at the top five cards of your library, then put them back in any
//	 order."
//
// Mirri's Guile's look-and-reorder over five cards; a library of fewer than five
// is looked at whole.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "613b824e-39d1-4c6e-a19c-c915abdc860b",
		Name:         "Index",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error { return effects.LookAtTop{N: 5}.Apply(ctx) })),
	})
}
