package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Taigam's Scheming — Sorcery {1}{U} (Khans of Tarkir):
//
//	"Surveil 5."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d9c57520-9827-4b8b-8f81-43cca4227c60",
		Name:         "Taigam's Scheming",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.Surveil{N: 5}.Apply(ctx)
		},
	})
}
