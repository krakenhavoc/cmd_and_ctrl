package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Selvala's Enforcer — Creature — Elf Warrior {3}{G}, 2/2 (Conspiracy, S82
// OLD batch 364, #5557):
//
//	"Parley — When this creature enters, each player reveals the top card of
//	 their library. For each nonland card revealed this way, put a +1/+1
//	 counter on this creature. Then each player draws a card."
//
// Cmdr2S49Parley with the counters as the reward, put on the Enforcer as the
// ability resolves (nothing happens to a creature that has left, but every
// player still draws).
//
// No simplification.
func init() {
	const label = "Selvala's Enforcer — each player reveals the top card of their library; put a +1/+1 counter on this creature for each nonland card; then each player draws a card"
	effects.Register(effects.Spec{
		OracleID:     "8ee2dd43-cda2-4cb7-9e4b-1946d5b7467e",
		Name:         "Selvala's Enforcer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.Cmdr2S49Parley(g, ctx.Source(), label, func(nonland int) error {
					if nonland == 0 {
						return nil
					}
					return effects.CounterOnThis{Kind: game.CounterPlusOne, N: nonland}.Apply(ctx)
				})
			}),
		},
	})
}
