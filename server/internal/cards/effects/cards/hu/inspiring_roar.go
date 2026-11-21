package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inspiring Roar — Sorcery {3}{W} (Aether Revolt,
// #5536):
//
//	"Put a +1/+1 counter on each creature you control."
//
// Basri's Solidarity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "96cf2810-35b8-437e-b609-f9765604b8cf",
		Name:         "Inspiring Roar",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B11PutCounterOnEachCreatureYouControl(ctx.Game, item)
		},
	})
}
