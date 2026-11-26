package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Selvala's Charge — Sorcery {4}{G} (Conspiracy, S82 OLD batch 364, #5557):
//
//	"Parley — Each player reveals the top card of their library. For each
//	 nonland card revealed this way, you create a 3/3 green Elephant
//	 creature token. Then each player draws a card."
//
// The parley body shared by Selvala, Explorer Returned and Storm Fleet
// Negotiator (Cmdr2S49Parley): every player reveals in turn order, the
// Elephants (all the caster's) come before the draws.
//
// No simplification.
func init() {
	const label = "Selvala's Charge — each player reveals the top card of their library; you create a 3/3 Elephant for each nonland card; then each player draws a card"
	effects.Register(effects.Spec{
		OracleID:     "7e6478c1-e5b3-4641-baa7-7cdeab042569",
		Name:         "Selvala's Charge",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Tokens: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.Cmdr2S49Parley(ctx.Game, ctx.Source(), label, func(nonland int) error {
				if nonland == 0 {
					return nil
				}
				return effects.CreateToken{Template: effects.TokenCard("3/3 green Elephant"), N: nonland}.Apply(ctx)
			})
		},
	})
}
