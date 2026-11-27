package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fade Away — Sorcery {2}{U} (Exodus, S82 OLD batch 742, #5626):
//
//	"For each creature, its controller sacrifices a permanent of their choice
//	 unless they pay {1}."
//
// Aether Barrier's ask, once for each creature on the battlefield as the spell
// resolves: the creature's controller pays {1} or sacrifices a permanent of
// their choice (a player with none sacrifices nothing).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f92c2d47-f2f7-43d9-acf3-0145bad1977d",
		Name:         "Fade Away",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB742FadeAway(ctx.Game, item)
		},
	})
}
