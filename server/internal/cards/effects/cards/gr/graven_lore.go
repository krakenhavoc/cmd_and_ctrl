package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Graven Lore — Snow Instant {3}{U}{U} (Kaldheim):
//
//	"Scry X, where X is the amount of {S} spent to cast this spell, then
//	 draw three cards."
//
// X is the snow mana that paid for the spell (CR 107.4h). The draw rides the
// scry's tail, so the three cards come after the player has put the scried
// cards back; with no snow mana spent there is nothing to scry (CR 701.22)
// and the cards are drawn at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "0d9df234-520c-40aa-a37d-070181cf695c",
		Name:           "Graven Lore",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsPoolManaOnly,
		WantsManaFrom:  game.ManaSourceSnow,
		Purpose:        game.Purpose{Draws: 3},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := ctx.ManaSpent().CountFrom(game.ManaSourceSnow)
			if n <= 0 {
				return effects.DrawCards{Player: item.Controller, N: 3}.Apply(ctx)
			}
			return effects.Scry{
				Player:   item.Controller,
				N:        n,
				ThenTail: ctx.TailN(effects.ThenDrawCards, 3),
			}.Apply(ctx)
		},
	})
}
