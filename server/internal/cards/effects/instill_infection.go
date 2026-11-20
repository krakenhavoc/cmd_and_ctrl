package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0040db73-ac11-4c15-a1ac-5a2c4c085f09",
		Name:         "Instill Infection",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      Harmful(TargetCreature("target creature")),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return putACounterOnTheTarget(game.CounterMinusOne)(ctx.Game, ctx.Item) }),
			DrawCards{N: 1},
		),
	})
}
