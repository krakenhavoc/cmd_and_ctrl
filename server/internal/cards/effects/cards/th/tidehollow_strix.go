package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tidehollow Strix — Artifact Creature — Bird {U}{B}, 2/1 (Shards of
// Alara, #5588):
//
//	"Flying
//	 Deathtouch"
//
// Two printed keywords: flying (CR 702.9b) and deathtouch (CR 702.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "19a39cc5-2134-4dee-932c-81389806c57b",
		Name:            "Tidehollow Strix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
	})
}
