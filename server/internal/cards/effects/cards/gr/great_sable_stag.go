package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Great Sable Stag — Creature — Elk {1}{G}{G}, 3/3 (Magic 2010, #5585):
//
//	"This spell can't be countered.
//	 Protection from blue and from black (This creature can't be blocked,
//	 targeted, dealt damage, or enchanted by anything blue or black.)"
//
// Shifting Ceratops' shape: the spell can't be countered (CR 101.2), and
// protection from blue and from black are two printed keywords
// (CR 702.16b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d6dbb83d-3215-4446-8480-8f1bf8cf2938",
		Name:            "Great Sable Stag",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{game.ProtectionFromColor("U"), game.ProtectionFromColor("B")},
	})
}
