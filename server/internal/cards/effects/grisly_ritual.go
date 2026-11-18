package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grisly Ritual — Sorcery {5}{B} (Innistrad: Crimson Vow):
//
//	"Destroy target creature or planeswalker. Create two Blood tokens."
func init() {
	Register(Spec{
		OracleID:     "56c78349-e4fa-48ba-ae8e-d2312ae6edc6",
		Name:         "Grisly Ritual",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 2},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := destroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return s82Vow03CreateBlood(2).Apply(ctx)
		},
	})
}
