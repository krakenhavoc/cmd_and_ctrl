package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tangle Spider — Creature — Spider {4}{G}{G} (Darksteel, #5605):
//
//	"Flash (You may cast this spell any time you could cast an instant.)
//	 Reach (This creature can block creatures with flying.)"
//
// Bounding Wolf's two printed keywords (CR 702.8, CR 702.17).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "597bd341-c5a0-47a3-9864-1fccab8a2a00",
		Name:            "Tangle Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "reach"},
	})
}
