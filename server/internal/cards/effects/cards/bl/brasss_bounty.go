package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brass's Bounty — Sorcery {6}{R} (Foundations, #4465):
//
//	"For each land you control, create a Treasure token."
//
// Avenger of Zendikar's land count with Treasures in place of Plants:
// the lands you control are counted as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "63b0538e-aa94-4e23-adf1-704cbf5bc3e5",
		Name:         "Brass's Bounty",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := effects.B02CountLandsControlledBy(ctx.Game, item.Controller)
			if n == 0 {
				return nil
			}
			return effects.CreateToken{Controller: item.Controller, Template: effects.TreasureToken(), N: n}.Apply(ctx)
		},
	})
}
