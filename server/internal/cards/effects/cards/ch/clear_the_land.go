package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clear the Land — Sorcery {2}{G} (Mercadian Masques, S82 OLD batch 697, #5619):
//
//	"Each player reveals the top five cards of their library, puts all land
//	 cards revealed this way onto the battlefield tapped, and exiles the rest."
//
// S82OldB697ClearTheLand: everyone reveals, the lands enter tapped as one
// batch under their owners' control, then the other revealed cards are exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4604c729-8d2d-4b2b-a95c-5245563637e6",
		Name:         "Clear the Land",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.S82OldB697ClearTheLand,
	})
}
