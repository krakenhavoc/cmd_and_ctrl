package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prologue to Phyresis — Instant {1}{U}: "Each opponent gets a poison counter.
// Draw a card." No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f661b58-e8ca-412c-bdac-af151162de12",
		Name:         "Prologue to Phyresis",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: effects.ResolveSteps(
			effects.S82OneB04PoisonStep(),
			effects.Instruction(func(ctx *effects.Context) error { return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx) }),
		),
	})
}
