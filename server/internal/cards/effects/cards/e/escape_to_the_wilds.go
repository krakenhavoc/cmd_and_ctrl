package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Escape to the Wilds — Sorcery {3}{R}{G} (Marvel Universe, #4460):
//
//	"Exile the top five cards of your library. You may play cards exiled
//	 this way until the end of your next turn.
//	 You may play an additional land this turn."
//
// Burning Curiosity's impulse exile (ExileTopNUntilYourNextTurn) with a
// count of five, then Explore's one-shot extra land drop for the rest of
// this turn. The extra drop is granted whether or not any exiled card is
// a land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "45a8c126-5396-4736-a37e-460a50706a98",
		Name:         "Escape to the Wilds",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{ExtraLandDrops: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.ExileTopNUntilYourNextTurn(ctx, 5); err != nil {
				return err
			}
			ctx.Game.GrantAdditionalLandPlayForEffect(item.Controller, 1)
			return nil
		},
	})
}
