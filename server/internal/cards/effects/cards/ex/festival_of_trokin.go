package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Festival of Trokin — Sorcery {W} (Portal Second Age, S82 older sets batch 735, #5625):
//
//	"You gain 2 life for each creature you control."
//
// Counted as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5a3cd2c-20ae-49a8-88fa-e8e3e7cc50db",
		Name:         "Festival of Trokin",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.GainLife{Amount: 2 * effects.B14CreaturesControlled(ctx.Game, item.Controller)}.Apply(ctx)
		},
	})
}
