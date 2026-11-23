package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Everdream — Instant {1}{U}:
//
//	"Draw a card.
//	Splice onto instant or sorcery {2}{U}"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b17d3c5-6085-416e-9afd-8ed0639ae7d2",
		Name:         "Everdream",
		Completeness: effects.CompletenessFull,
		Splice:       effects.SpliceOntoInstantOrSorcery("{2}{U}"),
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
