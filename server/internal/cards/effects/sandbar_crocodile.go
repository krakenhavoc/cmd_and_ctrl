package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sandbar Crocodile — Creature — Crocodile {4}{U}, 6/5 (Mirage, #5633):
//
//	"Phasing (This phases in or out before you untap during each of your
//	 untap steps. While it's phased out, it's treated as though it doesn't
//	 exist.)"
//
// Phasing is a printed keyword the engine honours on its own (CR 702.26a,
// game/phasing.go, ADR 0084): it phases out in its controller's next
// untap step and back in the one after, before that player untaps
// (CR 502.1), keeping its counters, damage and Auras, with no zone
// change. While it is phased out it is treated as though it doesn't
// exist (CR 702.26b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e11926bf-ca14-4e1e-aca5-9ffedb91868c",
		Name:            "Sandbar Crocodile",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordPhasing},
	})
}
