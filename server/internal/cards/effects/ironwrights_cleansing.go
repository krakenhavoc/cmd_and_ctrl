package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ironwright's Cleansing — Sorcery {2}{W} (Eldritch Moon, #5541):
//
//	"Exile target artifact or enchantment."
//
// Revoke Existence's text and body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "599faf5b-88fd-4373-b6ce-6047a67b30ea",
		Name:         "Ironwright's Cleansing",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ExileFirstTarget(ctx.Game, item)
		},
	})
}
