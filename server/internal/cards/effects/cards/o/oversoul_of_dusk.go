package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oversoul of Dusk — Creature — Spirit Avatar {G/W}{G/W}{G/W}{G/W}{G/W},
// 5/5 (Shadowmoor, #5591):
//
//	"Protection from blue, from black, and from red"
//
// Three protections from colours (CR 702.16b), each its own printed
// keyword, as on Animar's protection from white and from black.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "da790c72-63e7-4955-9435-d073767f9bb5",
		Name:            "Oversoul of Dusk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from blue", "protection from black", "protection from red"},
	})
}
