package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mass Appeal — {2}{U} Sorcery: "Draw a card for each Human you control."
//
// The Humans are counted as the spell resolves; a changeling counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aef067b4-cd9c-4f18-8c63-fbbc39b63376",
		Name:         "Mass Appeal",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := len(effects.PermanentsControlledByMatching(ctx.Game, ctx.Controller(), effects.OfCreatureType("Human")))
			return effects.DrawCards{N: n}.Apply(ctx)
		},
	})
}
