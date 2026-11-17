package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Altar's Light — Instant {2}{W}{W} (Mirrodin, #5606):
//
//	"Exile target artifact or enchantment."
//
// Exorcise's body on an artifact-or-enchantment clause. A target that
// is no longer legal as the spell resolves leaves it with nothing to do
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa9b6be2-b88c-4302-b7e2-faf25a60bcb9",
		Name:         "Altar's Light",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ExileFirstTarget(ctx.Game, item)
		},
	})
}
