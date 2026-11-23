package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dismiss — Instant {2}{U}{U} (Tempest, #5628):
//
//	"Counter target spell.
//	 Draw a card."
//
// Counterspell's counter (counterTheTargetSpell, CR 701.6a), then the
// draw (CR 608.2c). The target is a spell only, as printed: an ability
// on the stack can't be chosen. A spell that left the stack in response
// is an illegal target, so Dismiss doesn't resolve and nothing is drawn
// (CR 608.2b), as in paper.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c828a5e-10ae-4f63-86fd-2f160af43cc3",
		Name:         "Dismiss",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.CounterTheTargetSpell(ctx.Item, ctx) }),
			effects.DrawCards{N: 1},
		),
	})
}
