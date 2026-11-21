package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mountain Yeti — Creature — Yeti {2}{R}{R}, 3/3 (Legends, #5639):
//
//	"Mountainwalk (This creature can't be blocked as long as defending player controls a Mountain.)
//	 Protection from white"
//
// Mountainwalk is enforced at the block declaration (CR 702.14c), and
// protection from white refuses white spells' and white sources' targets
// (CR 702.16b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "375ea10b-da68-4c66-b3eb-f5113c09ea4b",
		Name:            "Mountain Yeti",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"mountainwalk", "protection from white"},
	})
}
