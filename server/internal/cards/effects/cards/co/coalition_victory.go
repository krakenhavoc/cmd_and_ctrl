package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coalition Victory — Sorcery {3}{W}{U}{B}{R}{G} (Invasion, S82 older sets batch 662, #5616):
//
//	"You win the game if you control a land of each basic land type and a
//	 creature of each color."
//
// The condition is read as the spell resolves (CR 608.2c), through the
// layers: a land has a basic land type by its subtypes, a creature has a
// color by its colors, and one creature of two colors counts for both. The
// win is WinTheGame's (CR 104.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9e25800-9ee7-40c9-b22d-611c7281c125",
		Name:         "Coalition Victory",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.OldB662CoalitionVictory(ctx.Game, item)
		},
	})
}
