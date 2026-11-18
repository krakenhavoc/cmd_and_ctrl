package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pointed Discussion — Sorcery {2}{B} (Innistrad: Crimson Vow):
//
//	"You draw two cards, lose 2 life, then create a Blood token."
func init() {
	Register(Spec{
		OracleID:     "aff9a568-2848-4614-a28b-b1ea84f7f0b7",
		Name:         "Pointed Discussion",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (DrawCards{N: 2}).Apply(ctx); err != nil {
				return err
			}
			if err := (GainLife{Amount: -2}).Apply(ctx); err != nil {
				return err
			}
			return s82Vow03CreateBlood(1).Apply(ctx)
		},
	})
}
