package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vedalken Outlander — Artifact Creature — Vedalken Scout {W}{U}, 2/2 (Conflux,
// #5587):
//
//	"Protection from red"
//
// Protection from red is a printed keyword, read by targeting, blocking,
// damage and attaching (CR 702.16a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "814a30dc-14e4-45cb-9320-b043e686b8ba",
		Name:            "Vedalken Outlander",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from red"},
	})
}
