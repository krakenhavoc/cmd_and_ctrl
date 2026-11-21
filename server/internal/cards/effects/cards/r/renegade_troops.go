package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Renegade Troops — Creature — Human Soldier {4}{R}, 4/2
// (Portal Three Kingdoms, #5622):
//
//	"Haste"
//
// Haste is an engine keyword carried as printed data: it can attack and
// tap the turn it arrives (CR 702.10b, CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "83a81308-4a15-49cd-b2f6-51431cca79e9",
		Name:            "Renegade Troops",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
