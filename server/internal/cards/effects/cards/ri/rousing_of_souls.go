package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rousing of Souls — Sorcery {2}{W} (Conspiracy, S82 OLD batch 364, #5557):
//
//	"Parley — Each player reveals the top card of their library. For each
//	 nonland card revealed this way, you create a 1/1 white Spirit creature
//	 token with flying. Then each player draws a card."
//
// Cmdr2S49Parley with the Spirits as the reward: every player reveals in turn
// order, the Spirits (all the caster's) come before the draws.
//
// No simplification.
func init() {
	const label = "Rousing of Souls — each player reveals the top card of their library; you create a 1/1 Spirit with flying for each nonland card; then each player draws a card"
	effects.Register(effects.Spec{
		OracleID:     "e5e42293-ce01-4872-be2a-a79ff4312992",
		Name:         "Rousing of Souls",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Tokens: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.Cmdr2S49Parley(ctx.Game, ctx.Source(), label, func(nonland int) error {
				if nonland == 0 {
					return nil
				}
				return effects.CreateToken{Template: effects.TokenCard("1/1 white Spirit with flying"), N: nonland}.Apply(ctx)
			})
		},
	})
}
