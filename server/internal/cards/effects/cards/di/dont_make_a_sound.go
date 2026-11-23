package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Don't Make a Sound — Instant {1}{U} (Duskmourn):
//
//	"Counter target spell unless its controller pays {2}. If they do,
//	 surveil 2."
//
// Assimilate Essence's shape: the pay-unless prompt (CounterUnlessPaid) is
// answered before anything else resolves, and the "if they do" is its keyed
// paid branch. The surveil is Don't Make a Sound's CONTROLLER's, not the
// payer's: the branch runs against a context bound to this spell.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f28a373b-4a8d-40a5-ace1-3bbd8489ad36",
		Name:         "Don't Make a Sound",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return effects.CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     "{2}",
				Question: "Don't Make a Sound — pay {2} or your spell is countered",
				Then:     dontMakeASoundPaid,
			}.Apply(ctx)
		},
	})
}

// dontMakeASoundPaid is "If they do, surveil 2", keyed. A key's branches
// replace the engine's own "counter that spell", so the decline is named
// here. An on-disk key: never renamed, never reused.
var dontMakeASoundPaid = effects.PayUnlessBranches("pay-unless/dont-make-a-sound-surveil",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error { return effects.Surveil{N: 2}.Apply(ctx) },
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		return ctx.Game.CounterIfStillOnStackForEffect(r.Guards)
	})
