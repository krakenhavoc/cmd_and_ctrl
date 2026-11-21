package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Instill Infection — Instant {3}{B} (Scars of Mirrodin, #5577):
//
//	"Put a -1/-1 counter on target creature.
//	 Draw a card."
//
// Fume Spitter's counter (putACounterOnTheTarget) on Honor's
// counter-then-draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0040db73-ac11-4c15-a1ac-5a2c4c085f09",
		Name:         "Instill Infection",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.Harmful(effects.TargetCreature("target creature")),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.PutACounterOnTheTarget(game.CounterMinusOne)(ctx.Game, ctx.Item)
			}),
			effects.DrawCards{N: 1},
		),
	})
}
