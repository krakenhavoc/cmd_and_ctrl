package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ray of Dissolution — Instant {2}{W} (Theros, S82 OLD b392, #5561):
//
//	"Destroy target enchantment. You gain 3 life."
//
// Break the Spell's target with Root Out's follow-on: the life is gained even
// if the enchantment survives; only a target that is illegal as the spell
// resolves stops the whole spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "509bfcbc-a59a-47c8-be78-29823fc010e3",
		Name:         "Ray of Dissolution",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, LifeGain: 3},
		Targets:      effects.TargetPermanent("target enchantment", effects.Enchantment()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DestroyTheTargetPermanent(ctx.Item, ctx) }),
			effects.GainLife{Amount: 3},
		),
	})
}
