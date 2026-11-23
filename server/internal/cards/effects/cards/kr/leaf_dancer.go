package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leaf Dancer — Creature — Centaur {1}{G}{G}, 2/2 (Odyssey, #5613):
//
//	"Forestwalk (This creature can't be blocked as long as defending player
//	 controls a Forest.)"
//
// Landwalk is read off the defending player's lands as blockers are
// declared (CR 702.14c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b2f92310-3725-44c2-8953-8b280437217e",
		Name:            "Leaf Dancer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"forestwalk"},
	})
}
