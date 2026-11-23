package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Greenwood Sentinel — Creature — Elf Scout {1}{G}, 2/2 (Foundations
// Jumpstart):
//
//	"Vigilance"
//
// A printed keyword. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8178c546-84f8-4563-a804-6489ec016660",
		Name:            "Greenwood Sentinel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
