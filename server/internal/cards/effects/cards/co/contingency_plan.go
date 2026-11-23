package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Contingency Plan — Sorcery {1}{U} (Eldritch Moon, #5541):
//
//	"Surveil 5. (Look at the top five cards of your library, then put any
//	 number of them into your graveyard and the rest on top of your
//	 library in any order.)"
//
// The surveil primitive (CR 701.25) for five cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "67bc4116-539e-4748-95e1-033df78173ec",
		Name:         "Contingency Plan",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.Surveil{N: 5}),
	})
}
