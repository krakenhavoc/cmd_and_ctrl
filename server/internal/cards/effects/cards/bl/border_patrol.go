package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Border Patrol — Creature — Human Nomad {4}{W}, 1/6 (Judgment):
//
//	"Vigilance"
//
// Vigilance is an engine keyword carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fa5e2575-51bb-448e-9aaf-0e0f5956ac98",
		Name:            "Border Patrol",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
