package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Squirming Mass — Creature — Horror {1}{B}, 1/1 (Urza's Destiny, #5621):
//
//	"Fear"
//
// One printed keyword, which the block check reads: only artifact
// creatures and black creatures can block it (CR 702.36b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2c526c61-6cdb-440c-b4c8-b6d1503cef73",
		Name:            "Squirming Mass",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"fear"},
	})
}
