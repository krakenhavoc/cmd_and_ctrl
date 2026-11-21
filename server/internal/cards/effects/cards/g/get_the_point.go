package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Get the Point — Instant {3}{B}{R} (Ravnica Allegiance, #5521):
//
//	"Destroy target creature. Scry 1."
//
// Murder with a scry 1. The destroy, then the scry as a second step
// (ResolveSteps). An illegal target counters the whole spell, scry
// included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57c704bc-8dff-4b90-b4e1-38d12e41b3d1",
		Name:         "Get the Point",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 1},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.ResolveSteps(effects.Instruction(effects.KineticOozeDestroy), effects.Scry{N: 1}),
	})
}
