package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fateful End — Instant {2}{R} (Theros Beyond Death, #5514):
//
//	"Fateful End deals 3 damage to any target. Scry 1."
//
// Lightning Strike's 3 damage to any target, then the scry; the scry
// waits for any prompt the damage asks (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5495ab6-081a-42be-b50c-f89f15a587c8",
		Name:         "Fateful End",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose: game.Purpose{
			Targets: game.ForTargets(effects.DamageToTarget(0, 3)),
			Answers: game.AnswerRemove,
			Scry:    1,
		},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DamageToFirstTargetStep(3)(ctx.Item, ctx) }),
			effects.Scry{N: 1},
		),
	})
}
