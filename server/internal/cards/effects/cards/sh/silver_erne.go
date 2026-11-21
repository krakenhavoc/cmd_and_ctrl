package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silver Erne — Creature — Bird {3}{U}, 2/2 (Ice Age, #5636):
//
//	"Flying, trample"
//
// Both are printed keywords: only creatures with flying or reach can
// block it (CR 702.9b), and damage beyond lethal to its blockers goes
// on to the player or permanent it attacks (CR 702.19b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6ca90451-861d-4aa8-93e7-3f3e02385530",
		Name:            "Silver Erne",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
	})
}
