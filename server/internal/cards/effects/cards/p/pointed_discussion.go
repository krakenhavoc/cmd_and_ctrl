package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pointed Discussion — Sorcery {2}{B} (Innistrad: Crimson Vow):
//
//	"You draw two cards, lose 2 life, then create a Blood token."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aff9a568-2848-4614-a28b-b1ea84f7f0b7",
		Name:         "Pointed Discussion",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.DrawCards{N: 2}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GainLife{Amount: -2}).Apply(ctx); err != nil {
				return err
			}
			return effects.S82Vow03CreateBlood(1).Apply(ctx)
		},
	})
}
