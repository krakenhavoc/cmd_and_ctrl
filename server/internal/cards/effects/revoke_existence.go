package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Revoke Existence — Sorcery {1}{W} (Scars of Mirrodin, #5577):
//
//	"Exile target artifact or enchantment."
//
// Altar's Light's text at sorcery speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "289e1a7c-b301-4395-baf8-e9a1c6d38aab",
		Name:         "Revoke Existence",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ExileFirstTarget(ctx.Game, item)
		},
	})
}
