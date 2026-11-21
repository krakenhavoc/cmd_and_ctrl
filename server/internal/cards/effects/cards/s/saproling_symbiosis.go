package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saproling Symbiosis — Sorcery {3}{G} (Dominaria Remastered, #5948):
//
//	"You may cast this spell as though it had flash if you pay {2} more to
//	 cast it.
//	 Create a 1/1 green Saproling creature token for each creature you
//	 control."
//
// The number of creatures is counted as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "69d22037-fbf7-44d2-81a7-7ce8e32f06ec",
		Name:          "Saproling Symbiosis",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Tokens: 1, CreatureTokens: 1},
		OptionalCosts: []game.AdditionalCost{effects.FlashForMore("{2}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := len(effects.MatchingBattlefield(ctx, effects.And(effects.YouControl(), effects.Creature())))
			if n == 0 {
				return nil
			}
			return effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: n}.Apply(ctx)
		},
	})
}
