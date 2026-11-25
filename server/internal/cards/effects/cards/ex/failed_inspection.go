package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Failed Inspection — Instant {2}{U}{U} (Kaladesh, #5539):
//
//	"Counter target spell. Draw a card, then discard a card."
//
// Counterspell's counter, then the loot: the discard waits for the draw
// (CR 608.2c) so the drawn card is a legal discard. With the target spell
// gone as this resolves the whole spell does nothing (CR 608.2b), no loot
// either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4fbf952-de3b-4cbd-848a-37041f337e06",
		Name:         "Failed Inspection",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.CounterTheTargetSpell(ctx.Item, ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 1, 1, "")
			}),
		),
	})
}
