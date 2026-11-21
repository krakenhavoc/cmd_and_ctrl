package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Giant Scorpion — Creature — Scorpion {2}{B}, 1/3 (Zendikar, #5583):
//
//	"Deathtouch (Any amount of damage this deals to a creature is
//	 enough to destroy it.)"
//
// Deathtouch is the printed keyword (CR 702.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "557485d9-7537-4297-a589-c29d0f1ecd98",
		Name:            "Giant Scorpion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
