package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dwarven Grunt — Creature — Dwarf {R}, 1/1 (Odyssey, #5613):
//
//	"Mountainwalk (This creature can't be blocked as long as defending player
//	 controls a Mountain.)"
//
// Landwalk is read off the defending player's lands as blockers are
// declared (CR 702.14c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6da00e2d-4287-48eb-8abe-d153318381fd",
		Name:            "Dwarven Grunt",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"mountainwalk"},
	})
}
