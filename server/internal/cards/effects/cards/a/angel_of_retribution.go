package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angel of Retribution — Creature — Angel {6}{W}, 5/5 (Torment, #5612):
//
//	"Flying, first strike"
//
// Two printed keywords the engine enforces: flying in block legality
// (CR 702.9b) and first strike in the combat damage steps (CR 702.7b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "248dcf69-b99e-4a71-b9b3-7fd865f873ab",
		Name:            "Angel of Retribution",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
	})
}
