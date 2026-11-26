package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rise of Eagles — Sorcery {4}{U}{U} (Journey into Nyx, S82 OLD, #5558):
//
//	"Create two 2/2 blue Bird enchantment creature tokens with flying. Scry 1."
//
// The tokens are enchantments as well as creatures, so they are constellation
// triggers for the enchantment-enters watchers. The scry follows the tokens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8fc154ff-f2a2-4089-a504-f483f545606e",
		Name:         "Rise of Eagles",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2, Scry: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.CreateToken{Template: effects.TokenCard("2/2 blue Bird enchantment with flying"), N: 2}).Apply(ctx); err != nil {
				return err
			}
			return effects.Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
