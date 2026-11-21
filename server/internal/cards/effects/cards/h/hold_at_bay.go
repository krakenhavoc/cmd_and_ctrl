package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hold at Bay — Instant {1}{W} (Born of the Gods,
// #5559):
//
//	"Prevent the next 7 damage that would be dealt to any target this turn."
//
// Aven Redeemer's shield as a spell (PreventNextDamage on the target
// still legal as it resolves).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a57b84e4-951c-4475-bb34-b68815858f87",
		Name:         "Hold at Bay",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPrevent},
		Targets:      effects.TargetAny(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				return effects.PreventNextDamage{Target: t.ID, Amount: 7, Label: "Hold at Bay: prevent the next 7 damage"}.Apply(ctx)
			}
			return nil
		},
	})
}
