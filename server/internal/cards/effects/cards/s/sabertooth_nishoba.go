package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sabertooth Nishoba — Creature — Cat Beast Warrior {4}{G}{W}, 5/5 (Invasion, S82 #5616):
//
//	"Trample, protection from blue and from red"
//
// Trample and two protections, each a printed keyword (CR 702.19, 702.16).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ce82aa79-bb50-47bc-84f5-eae30249c7a5",
		Name:            "Sabertooth Nishoba",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "protection from blue", "protection from red"},
	})
}
