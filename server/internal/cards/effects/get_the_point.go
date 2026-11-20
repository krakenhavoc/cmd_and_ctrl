package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "57c704bc-8dff-4b90-b4e1-38d12e41b3d1",
		Name:         "Get the Point",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 1},
		Targets:      TargetCreature("target creature"),
		OnResolve:    ResolveSteps(Instruction(kineticOozeDestroy), Scry{N: 1}),
	})
}
