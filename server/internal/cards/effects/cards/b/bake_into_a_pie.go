package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bake into a Pie — Instant {2}{B}{B} (Foundations, #4465):
//
//	"Destroy target creature. Create a Food token."
//
// The destroy, then the shared Food template. The Food waits for any
// prompt the destroy asks (CR 608.2c), and it is made even when the
// creature survives (indestructible, a regeneration shield): only a
// target that is illegal as the spell resolves stops the whole spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b9ec782-0ba1-41f1-bc39-d3302494ecb3",
		Name:         "Bake into a Pie",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DestroyTheTargetPermanent(ctx.Item, ctx) }),
			effects.CreateToken{Template: effects.FoodToken(), N: 1},
		),
	})
}
