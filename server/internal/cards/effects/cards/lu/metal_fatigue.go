package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Metal Fatigue — Instant {2}{W} (Darksteel, #5605):
//
//	"Tap all artifacts."
//
// Ensnare's tap-all over artifacts (dstB2TapAllArtifacts), every
// player's included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f817352a-908a-4cb5-8e10-feeb770540fc",
		Name:         "Metal Fatigue",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.Instruction(effects.DstB2TapAllArtifacts)),
	})
}
