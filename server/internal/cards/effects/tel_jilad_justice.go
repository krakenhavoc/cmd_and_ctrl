package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6c853b95-3d98-44fd-b05a-02d05d5eb0b8",
		Name:         "Tel-Jilad Justice",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 2},
		Targets:      TargetPermanent("target artifact", Artifact()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := destroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return Scry{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
