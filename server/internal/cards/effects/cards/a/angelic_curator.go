package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angelic Curator — Creature — Angel Spirit {1}{W}, 1/1 (Urza's Legacy,
// #5623):
//
//	"Flying, protection from artifacts"
//
// Both are printed keywords: flying limits its blockers (CR 702.9b), and
// protection from artifacts stops artifact spells and abilities from
// artifact sources targeting it and prevents damage from artifact sources
// (CR 702.16b, CR 702.16e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cca29a9b-794f-4712-8e6c-36ce3da9cb8b",
		Name:            "Angelic Curator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from artifacts"},
	})
}
