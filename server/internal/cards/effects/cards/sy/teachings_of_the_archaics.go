package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Teachings of the Archaics — Sorcery — Lesson {2}{U} (Strixhaven: School of
// Mages, #5502):
//
//	"If an opponent has more cards in hand than you, draw two cards. Draw
//	 three cards instead if an opponent has at least four more cards in hand
//	 than you."
//
// The comparison is one opponent against you (OldB42OpponentHandLead: the
// biggest lead any single opponent still in the game has), read as the spell
// resolves, when it is no longer in your hand (CR 608.2h). A tie or a smaller
// hand draws nothing. At a lead of four or more the replacement text makes it
// three cards, not five.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6ac3abba-cf55-4d97-a253-bce45972b050",
		Name:         "Teachings of the Archaics",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			lead := effects.OldB42OpponentHandLead(ctx.Game, ctx.Controller())
			switch {
			case lead >= 4:
				return effects.DrawCards{N: 3}.Apply(ctx)
			case lead > 0:
				return effects.DrawCards{N: 2}.Apply(ctx)
			}
			return nil
		},
	})
}
