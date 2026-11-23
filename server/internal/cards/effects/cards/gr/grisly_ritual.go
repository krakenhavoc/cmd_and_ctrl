package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grisly Ritual — Sorcery {5}{B} (Innistrad: Crimson Vow):
//
//	"Destroy target creature or planeswalker. Create two Blood tokens."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56c78349-e4fa-48ba-ae8e-d2312ae6edc6",
		Name:         "Grisly Ritual",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 2},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return effects.S82Vow03CreateBlood(2).Apply(ctx)
		},
	})
}
