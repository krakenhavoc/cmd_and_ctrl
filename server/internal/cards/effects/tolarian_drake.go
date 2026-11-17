package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tolarian Drake — Creature — Drake {2}{U}, 2/4 (Weatherlight, #5629):
//
//	"Flying
//	 Phasing (This phases in or out before you untap during each of your
//	 untap steps. While it's phased out, it's treated as though it doesn't
//	 exist.)"
//
// Teferi's Drake's two keywords on a 2/4 body. Phasing is the engine's
// own (CR 702.26a, game/phasing.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d999f2a3-5cb4-4f12-a79a-c5db0a300620",
		Name:            "Tolarian Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordPhasing},
	})
}
