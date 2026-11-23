package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pact of the Titan — Instant {0}, red (Future Sight, #5594):
//
//	"Create a 4/4 red Giant creature token.
//	 At the beginning of your next upkeep, pay {4}{R}. If you don't, you
//	 lose the game."
//
// A token at instant speed, then Pact of Negation's upkeep debt. The card
// is red by its colour indicator, which rides game.Card.Colors from
// Scryfall.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf4959bf-92ae-4c9b-b631-724ce100cbb8",
		Name:         "Pact of the Titan",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("4/4 red Giant"), N: 1}).Apply(ctx); err != nil {
				return err
			}
			return effects.FutB1OwePact(ctx, "Pact of the Titan", "{4}{R}")
		},
	})
}
