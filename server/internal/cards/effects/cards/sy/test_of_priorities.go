package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Test of Priorities — Instant {1}{U} (Star Trek):
//
//	Choose one —
//	• Power to Engines — Surveil 2, then draw a card.
//	• Power to Life Support — Exile target creature you control. Return
//	  that card to the battlefield under its owner's control at the
//	  beginning of the next end step.
//
// "Power to Engines" and "Power to Life Support" are mode names with no
// rules meaning. The first bullet is Prismari Charm's: the draw rides the
// surveil's continuation, so the card drawn comes from the library the
// player arranged. The second is Voidwalk's delayed blink on one creature
// the caster controls (a CR 603.7 delayed trigger returning it under its
// owner's control at the next end step).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd24de4f-ce88-4fbb-9e8d-93ea7c260b4c",
		Name:         "Test of Priorities",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Power to Engines — Surveil 2, then draw a card."), game.Purpose{Draws: 1}),
			effects.Mode("Power to Life Support — Exile target creature you control. Return that card to the battlefield under its owner's control at the beginning of the next end step.",
				effects.TargetCreature("target creature you control", effects.YouControl())),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			switch {
			case ctx.HasMode(0):
				return effects.Surveil{Player: item.Controller, N: 2, ThenTail: ctx.Tail(effects.ThenDrawACard)}.Apply(ctx)
			case ctx.HasMode(1):
				var ids []uuid.UUID
				for _, t := range effects.OptionTargets(ctx, 1) {
					if t.Kind == game.TargetCard {
						ids = append(ids, t.ID)
					}
				}
				return effects.ExileCardsThenScheduleReturn(ctx, ids, "Test of Priorities — return the exiled creature")
			}
			return nil
		},
	})
}
