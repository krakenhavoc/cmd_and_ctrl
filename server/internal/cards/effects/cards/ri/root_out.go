package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "bec03edc-261a-4b02-be37-2006f657e905",
		Name:         "Root Out",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DestroyTheTargetPermanent(ctx.Item, ctx) }),
			effects.Investigate{N: 1},
		),
	})
}
