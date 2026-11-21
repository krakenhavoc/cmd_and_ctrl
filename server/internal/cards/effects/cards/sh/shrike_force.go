package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shrike Force — Creature — Bird Knight {2}{W}, 1/3 (Bloomburrow, #4467):
//
//	"Flying, double strike, vigilance"
//
// Serra Angel's shape: three printed keywords and nothing else, fed to
// the layer 6 keyword static through PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d014b02f-a8b8-4b97-96e0-f588792069f0",
		Name:            "Shrike Force",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "double strike", "vigilance"},
	})
}
