package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ugin's Insight — Sorcery {3}{U}{U} (Modern Horizons 3 Commander, #5948):
//
//	"Scry X, where X is the greatest mana value among permanents you
//	 control, then draw three cards."
//
// X is read as the spell resolves, over the permanents its controller
// controls then (a token or a land is mana value 0, a copy has the copied
// mana value, an X in a cost counts as 0 on the battlefield, CR 202.3e).
// The scry finishes before the draw (Preordain's shape, a registered tail), so
// the cards left on top are the ones drawn. With X = 0 the scry does nothing
// and the three cards are still drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad7b9db0-e061-4360-9b81-0408d18d971d",
		Name:         "Ugin's Insight",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			greatest := 0
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == ctx.Controller() {
					if mv := c.ManaValue(); mv > greatest {
						greatest = mv
					}
				}
			}
			return effects.Scry{
				Player:   ctx.Controller(),
				N:        greatest,
				ThenTail: ctx.TailN(effects.ThenDrawCards, 3),
			}.Apply(ctx)
		},
	})
}
