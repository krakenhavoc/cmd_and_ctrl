package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Expose to Daylight — Instant {2}{W} (Ravnica Allegiance, #5521):
//
//	"Destroy target artifact or enchantment. Scry 1."
//
// Disenchant with a scry 1. The destroy, then the scry as a second step
// (ResolveSteps). An illegal target counters the whole spell, scry
// included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8195fcd-1858-4f23-aa73-b654d25035aa",
		Name:         "Expose to Daylight",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 1},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve:    effects.ResolveSteps(effects.Instruction(effects.KineticOozeDestroy), effects.Scry{N: 1}),
	})
}
