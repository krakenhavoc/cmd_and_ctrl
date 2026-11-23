package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mountain Bandit — Creature — Human Soldier Rogue {R}, 1/1
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
		OracleID:        "e3638c54-e494-47d2-b614-94c9d9add85a",
		Name:            "Mountain Bandit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
