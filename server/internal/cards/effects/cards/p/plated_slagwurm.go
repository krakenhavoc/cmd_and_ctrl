package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plated Slagwurm — Creature — Wurm {4}{G}{G}{G}, 8/8 (Mirrodin, #5606):
//
//	"Hexproof (This creature can't be the target of spells or abilities
//	 your opponents control.)"
//
// Hexproof is a printed keyword targeting honours (CR 702.11b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4aec7624-e406-45a4-b2b6-2e8d29f6268a",
		Name:            "Plated Slagwurm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
	})
}
