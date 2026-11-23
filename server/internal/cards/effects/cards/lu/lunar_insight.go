package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lunar Insight — Sorcery {2}{U} (Foundations, #4465):
//
//	"Draw a card for each different mana value among nonland permanents
//	 you control."
//
// The count is taken as the spell resolves: the distinct mana values of
// the nonland permanents its controller controls, a token that is not a
// copy counting as mana value 0 and an {X} in a mana cost as 0 on the
// battlefield (the ruling, CR 202.3e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56b4a2f7-2d35-4d05-8d26-4f9f16c5f9f5",
		Name:         "Lunar Insight",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: item.Controller, N: effects.StdDifferentManaValuesAmongYourNonlands(ctx.Game, item.Controller)}.Apply(ctx)
		},
	})
}
