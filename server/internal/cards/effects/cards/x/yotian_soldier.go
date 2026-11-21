package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yotian Soldier — Artifact Creature — Soldier {3}, 1/4 (Antiquities, #5640):
//
//	"Vigilance"
//
// One printed keyword: it rides PrintedKeywords, and attacking does not
// tap the Soldier.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "645571f1-a775-4834-8b79-8852d38c9587",
		Name:            "Yotian Soldier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
