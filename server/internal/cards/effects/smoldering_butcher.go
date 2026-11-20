package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Smoldering Butcher — Creature — Elemental Warrior {3}{B}, 4/2 (Eventide, #5590):
//
//	"Wither (This deals damage to creatures in the form of -1/-1 counters.)"
//
// The printed keyword (CR 702.80).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b0f56eef-1abc-437e-8e2e-c079731fe783",
		Name:            "Smoldering Butcher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordWither},
	})
}
