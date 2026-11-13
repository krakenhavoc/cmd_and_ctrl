package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teferi's Drake — Creature — Drake {2}{U}, 3/2 (Mirage, #5633):
//
//	"Flying
//	 Phasing (This phases in or out before you untap during each of your
//	 untap steps. While it's phased out, it's treated as though it doesn't
//	 exist.)"
//
// Both are printed keywords. Phasing is the engine's own (CR 702.26a,
// game/phasing.go, ADR 0084), as on Sandbar Crocodile: out in its
// controller's next untap step, back in the one after (CR 502.1).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3605c5ac-3b35-41a5-9266-286fb1d11716",
		Name:            "Teferi's Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordPhasing},
	})
}
