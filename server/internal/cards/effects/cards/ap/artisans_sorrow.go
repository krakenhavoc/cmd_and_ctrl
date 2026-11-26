package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Artisan's Sorrow — Instant {3}{G} (Theros, S82 OLD b392, #5561):
//
//	"Destroy target artifact or enchantment. Scry 2."
//
// Root Out's target and destroy with a scry in place of the Clue; the scry
// happens even if the permanent survives (indestructible).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "115fd5ac-460b-49a3-83ed-9d56a7aecb14",
		Name:         "Artisan's Sorrow",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 2},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DestroyTheTargetPermanent(ctx.Item, ctx) }),
			effects.Instruction(func(ctx *effects.Context) error { return effects.Scry{Player: ctx.Controller(), N: 2}.Apply(ctx) }),
		),
	})
}
