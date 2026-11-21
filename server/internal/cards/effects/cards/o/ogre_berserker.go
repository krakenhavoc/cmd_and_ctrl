package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ogre Berserker — Creature — Ogre Berserker {4}{R}, 4/2 (Portal Second Age,
// #5625):
//
//	"Haste"
//
// Haste is an engine keyword carried as printed data: it can attack the
// turn it arrives (CR 702.10b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9e6f76dc-b7db-49b6-a61e-3495b81e61d7",
		Name:            "Ogre Berserker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
