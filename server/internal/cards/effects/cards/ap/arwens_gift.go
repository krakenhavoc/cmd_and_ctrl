package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arwen's Gift — Sorcery {3}{U} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"This spell costs {1} less to cast if you control two or more legendary
//	 creatures.
//	 Scry 2, then draw two cards."
//
// The discount is a self cost modifier (ADR 0048) over the legendary creatures
// you control, read through the layers as the spell is priced. The draw is
// Preordain's "scry, then" tail: it waits for the scry to be answered, so the
// cards kept on top are the cards drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "13ac12db-d04b-4211-90df-886e1c597e8f",
		Name:         "Arwen's Gift",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2, Draws: 2, Answers: game.AnswerValue},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(1, "This spell costs {1} less to cast if you control two or more legendary creatures.",
				func(q game.CostQuery) bool {
					return effects.PermanentsYouControl(effects.And(effects.Creature(), effects.Legendary()))(q) >= 2
				}),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.Scry{N: 2, ThenTail: ctx.TailN(effects.ThenDrawCards, 2)}.Apply(ctx)
		},
	})
}
