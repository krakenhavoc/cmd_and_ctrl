package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tezzeret's Ambition — Sorcery {3}{U}{U} (Kaladesh, #5539):
//
//	"Draw three cards. If you control no artifacts, discard a card."
//
// The discard waits for the draw (CR 608.2c), as every "draw, then" clause
// does, and "if you control no artifacts" is read after the three cards are
// drawn, through the layered view (a drawn card is not a permanent, so the
// draw cannot change it). With an artifact the spell stops after the draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91486e70-94e4-431f-8899-fef019bb206e",
		Name:         "Tezzeret's Ambition",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.DrawNThenForEffect(ctx.Controller(), 3, game.DrawThen{
				Ref: effects.S82OldB259DiscardUnlessArtifactRef, Player: ctx.Controller(), Source: ctx.Source(), N: 1,
			})
		},
	})
}
