package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Summer Bloom — Sorcery {1}{G} (S82 GAP, #5948):
//
//	"You may play up to three additional lands this turn."
//
// Till and Tend's land-play grant, three times (CR 305.2).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5df4597-1647-4ac2-bdb3-a517598d1431",
		Name:         "Summer Bloom",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{ExtraLandDrops: 3},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ctx.Game.GrantAdditionalLandPlayForEffect(item.Controller, 3)
			return nil
		},
	})
}
