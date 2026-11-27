package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lone Wolf — Creature — Wolf {2}{G}, 2/2 (Eighth Edition, #5607):
//
//	"You may have this creature assign its combat damage as though it
//	 weren't blocked."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f77dfcb6-72b0-4d50-8177-6f6c57a034b9",
		Name:         "Lone Wolf",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.S82OldB614AssignAsThoughUnblocked()},
	})
}
