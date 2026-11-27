package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Broken Dam — Sorcery {U}
// (Portal Three Kingdoms, #5622):
//
//	"Tap one or two target creatures without horsemanship."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75fffc1b-db65-4c6b-9503-db50fccb837c",
		Name:         "Broken Dam",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("one or two target creatures without horsemanship", effects.WithoutKeyword("horsemanship")).WithCount(1, 2),
		OnResolve:    effects.ResolveSteps(effects.Instruction(effects.OldB711TapTheTargets)),
	})
}
