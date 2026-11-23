package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goliath Beetle — Creature — Insect {2}{G}, 3/1 (Urza's Destiny, #5621):
//
//	"Trample"
//
// One printed keyword, which the combat damage step reads: damage
// beyond what its blockers need for lethal goes to the player it
// attacks (CR 702.19b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "86ab4400-fbdd-4c18-a893-441286a9d7d0",
		Name:            "Goliath Beetle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
