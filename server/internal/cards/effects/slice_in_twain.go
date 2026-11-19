package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slice in Twain — Instant {2}{G}{G} (Scars of Mirrodin, #5577):
//
//	"Destroy target artifact or enchantment.
//	 Draw a card."
//
// Break Asunder's destroy, then a draw. If the target is gone by
// resolution the whole spell does nothing, the draw included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a96d0065-b8c1-4843-8752-dce60bf83545",
		Name:         "Slice in Twain",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := destroyFirstLegalCardTarget(ctx.Game, item); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
