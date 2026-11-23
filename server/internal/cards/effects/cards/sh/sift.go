package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sift — Sorcery {3}{U} (Stronghold, #5627):
//
//	"Draw three cards, then discard a card."
//
// Focus the Mind's shared drawThenDiscard body at three and one, without
// its discount: the three cards are drawn first, so any of them can be
// the one discarded, and the caster picks it (CR 701.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e060984-2015-4dc4-b53d-ab6bf2abdacc",
		Name:         "Sift",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3, Discards: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 3, 1, "")
		},
	})
}
