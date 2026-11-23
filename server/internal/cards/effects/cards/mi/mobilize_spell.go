package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mobilize — Sorcery {G} (Portal, #5630):
//
//	"Untap all creatures you control."
//
// Vitalize's body as a sorcery (b34UntapAllCreaturesYouControl).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7416c8db-ec33-4d92-866c-b7fdc69c8b41",
		Name:         "Mobilize",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B34UntapAllCreaturesYouControl(ctx.Game, item)
		},
	})
}
