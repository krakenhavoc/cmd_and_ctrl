package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Contentious Plan — Sorcery {1}{U} (War of the Spark, #5520):
//
//	"Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
//	 Draw a card."
//
// Steady Progress's body: the draw waits for the proliferate answer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fc6de19-9d04-422b-90bc-eee2f9f8c84c",
		Name:         "Contentious Plan",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.Proliferate{ThenTail: ctx.TailN(effects.ThenDrawCards, 1)}.Apply(ctx)
		},
	})
}
