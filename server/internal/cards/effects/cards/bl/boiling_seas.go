package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boiling Seas — Sorcery {3}{R} (Portal, S82 OLD batch 784, #5630):
//
//	"Destroy all Islands."
//
// Boil's body as a sorcery: every land with the Island subtype, basic or not,
// whoever controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7421b711-81a4-4042-b024-f55bf9ff203f",
		Name:         "Boiling Seas",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Subtype("Island")}.Apply(ctx)
		},
	})
}
