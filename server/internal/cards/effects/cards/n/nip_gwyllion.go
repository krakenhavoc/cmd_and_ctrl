package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nip Gwyllion — Creature — Hag {W/B}, 1/1 (Eventide, #5590):
//
//	"Lifelink (Damage dealt by this creature also causes you to gain that
//	 much life.)"
//
// The printed keyword (CR 702.15).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cd53f7cf-e4da-4e93-8550-60556300b56c",
		Name:            "Nip Gwyllion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
