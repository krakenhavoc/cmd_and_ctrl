package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rubble Reading — Sorcery {3}{R} (Ravnica Allegiance, #5521):
//
//	"Destroy target land. Scry 2."
//
// Stone Rain with a scry 2. The destroy, then the scry as a second step
// (ResolveSteps). An illegal target counters the whole spell, scry
// included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "894d96b1-c82a-4141-a2f3-76da505bd5e5",
		Name:         "Rubble Reading",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 2},
		Targets:      TargetPermanent("target land", Land()),
		OnResolve:    ResolveSteps(Instruction(kineticOozeDestroy), Scry{N: 2}),
	})
}
