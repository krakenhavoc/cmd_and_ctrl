package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vampire's Kiss — Sorcery {1}{B} (Innistrad: Crimson Vow):
//
//	"Target player loses 2 life and you gain 2 life. Create two Blood tokens."
func init() {
	drain := effects.TargetPlayerLosesLifeYouGain(2)
	effects.Register(effects.Spec{
		OracleID:     "105cea6e-18e9-4dec-a533-0dca4d41320b",
		Name:         "Vampire's Kiss",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      game.Purpose{Tokens: 2},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := drain(item, ctx); err != nil {
				return err
			}
			return effects.S82Vow03CreateBlood(2).Apply(ctx)
		},
	})
}
