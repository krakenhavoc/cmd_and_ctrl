package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Specter's Wail — Sorcery {1}{B} (Mercadian Masques, S82 OLD batch 697, #5619):
//
//	"Target player discards a card at random."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af32ce11-c026-4135-b1e3-10c6aa15c72c",
		Name:         "Specter's Wail",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve:    effects.ResolveSteps(effects.TargetPlayerDiscardsAtRandom{N: 1}),
	})
}
