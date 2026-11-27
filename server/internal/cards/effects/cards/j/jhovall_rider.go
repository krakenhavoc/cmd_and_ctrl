package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jhovall Rider — Creature — Human Rebel {4}{W}, 3/3 (Mercadian Masques, S82
// OLD b684, #5619):
//
//	"Trample"
//
// Trample is a printed keyword (CR 702.19a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "84750c9a-ddf7-4ee1-b93b-3dc9ab73e4d5",
		Name:            "Jhovall Rider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
