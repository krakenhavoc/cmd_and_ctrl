package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sudden Insight — Instant {4}{U}{U} (Adventures in the Forgotten Realms):
//
//	"Draw a card for each different mana value among nonland cards in your
//	 graveyard."
//
// Counted as the spell resolves. A card in a graveyard has its printed mana
// value (X is 0, CR 202.3e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8e0b605f-2095-4189-8c71-9d1ac93db42e",
		Name:         "Sudden Insight",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
			if p == nil || p.Graveyard == nil {
				return nil
			}
			seen := map[int]bool{}
			for _, c := range p.Graveyard.Cards {
				if !c.IsLand() {
					seen[c.ManaValue()] = true
				}
			}
			return effects.DrawCards{N: len(seen)}.Apply(ctx)
		},
	})
}
