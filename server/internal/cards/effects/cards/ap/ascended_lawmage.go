package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ascended Lawmage — Creature — Vedalken Wizard {2}{W}{U}, 3/2:
//
//	"Flying
//	 Hexproof (This creature can't be the target of spells or abilities
//	 your opponents control.)"
//
// Two printed keywords the engine reads: flying at the block declaration
// (CR 702.9b) and hexproof at targeting (CR 702.11b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4317f4f1-d339-4940-b46e-659641035595",
		Name:            "Ascended Lawmage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "hexproof"},
	})
}
