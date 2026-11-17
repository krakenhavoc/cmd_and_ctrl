package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prologue to Phyresis — Instant {1}{U}: "Each opponent gets a poison counter.
// Draw a card." No simplification.
func init() {
	Register(Spec{
		OracleID:     "3f661b58-e8ca-412c-bdac-af151162de12",
		Name:         "Prologue to Phyresis",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: ResolveSteps(
			s82OneB04PoisonStep(),
			Instruction(func(ctx *Context) error { return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx) }),
		),
	})
}
