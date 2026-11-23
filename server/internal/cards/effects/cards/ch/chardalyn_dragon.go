package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chardalyn Dragon — Artifact Creature — Dragon {6}, 4/4 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"Flying"
//
// A vanilla flier: its only ability is the keyword, from the printed
// keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0bf1f2c1-2148-4302-a6e7-d509a005a85f",
		Name:            "Chardalyn Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
