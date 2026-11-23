package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tel-Jilad Justice — Instant {1}{G} (Fifth Dawn, #5604):
//
//	"Destroy target artifact. Scry 2. (Look at the top two cards of
//	 your library, then put any number of them on the bottom and the
//	 rest on top in any order.)"
//
// Shatter, then the scry, which happens whenever the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c853b95-3d98-44fd-b05a-02d05d5eb0b8",
		Name:         "Tel-Jilad Justice",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 2},
		Targets:      effects.TargetPermanent("target artifact", effects.Artifact()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return effects.Scry{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
