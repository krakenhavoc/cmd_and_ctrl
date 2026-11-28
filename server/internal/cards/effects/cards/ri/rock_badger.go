package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rock Badger — Creature — Badger Beast {4}{R}, 3/3 (Mercadian Masques, S82
// OLD b683, #5619):
//
//	"Mountainwalk (This creature can't be blocked as long as defending player
//	 controls a Mountain.)"
//
// Mountainwalk is a printed keyword (CR 702.14a), judged as blockers are
// declared against the defending player's lands.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1f5ac983-a7a5-4b9d-a25c-727a7de6c465",
		Name:            "Rock Badger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"mountainwalk"},
	})
}
