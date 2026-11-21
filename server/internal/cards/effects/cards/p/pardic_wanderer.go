package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pardic Wanderer — Artifact Creature — Golem {6}, 5/5 (Dominaria, #5530):
//
//	"Trample"
//
// One printed keyword, read by combat damage assignment (CR 702.19b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "82ff4500-b118-4ce7-9f7c-92661328b242",
		Name:            "Pardic Wanderer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
