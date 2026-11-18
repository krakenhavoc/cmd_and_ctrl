package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Root Out — Sorcery {2}{G}:
//
//	"Destroy target artifact or enchantment.
//	 Investigate."
//
// Abolish's target and destroy, then Bake into a Pie's follow-on: the
// Clue is made even if the permanent survives (indestructible), and only
// a target that is illegal as the spell resolves stops the whole spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bec03edc-261a-4b02-be37-2006f657e905",
		Name:         "Root Out",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return destroyTheTargetPermanent(ctx.Item, ctx) }),
			Investigate{N: 1},
		),
	})
}
