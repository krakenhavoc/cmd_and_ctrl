package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bone to Ash — Instant {2}{U}{U} (Dark Ascension, #5571):
//
//	"Counter target creature spell.
//	 Draw a card."
//
// Dismiss narrowed to a creature spell. A target that left the stack in
// response counters the whole spell on resolution (CR 608.2b), so no card
// is drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "31be5bf4-9950-456c-8365-74b48e132ef5",
		Name:         "Bone to Ash",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetSpell("target creature spell", effects.Creature()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.CounterTheTargetSpell(ctx.Item, ctx) }),
			effects.DrawCards{N: 1},
		),
	})
}
