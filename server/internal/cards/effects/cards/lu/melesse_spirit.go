package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Melesse Spirit — Creature — Angel Spirit {3}{W}{W}, 3/3 (Mirage, S82 OLD
// b806):
//
//	"Flying, protection from black"
//
// Both are engine keywords carried as printed data; protection is one quality
// token (CR 702.16).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1c4bd85b-03c2-4235-aea1-76c48520adf3",
		Name:            "Melesse Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.ProtectionFromColor("B")},
	})
}
