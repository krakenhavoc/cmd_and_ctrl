package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Titania's Boon — Sorcery {3}{G} (Urza's Saga, S82 OLD b724):
//
//	Put a +1/+1 counter on each creature you control.
//
// Basri's Solidarity's body: every creature the caster controls as the spell resolves gets a counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0bf4c01d-3cb0-47be-b84a-6a399bc2d3fc",
		Name:         "Titania's Boon",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B11PutCounterOnEachCreatureYouControl(ctx.Game, item)
		},
	})
}
