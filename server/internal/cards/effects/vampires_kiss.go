package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vampire's Kiss — Sorcery {1}{B} (Innistrad: Crimson Vow):
//
//	"Target player loses 2 life and you gain 2 life. Create two Blood tokens."
func init() {
	drain := targetPlayerLosesLifeYouGain(2)
	Register(Spec{
		OracleID:     "105cea6e-18e9-4dec-a533-0dca4d41320b",
		Name:         "Vampire's Kiss",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		Purpose:      game.Purpose{Tokens: 2},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := drain(item, ctx); err != nil {
				return err
			}
			return s82Vow03CreateBlood(2).Apply(ctx)
		},
	})
}
