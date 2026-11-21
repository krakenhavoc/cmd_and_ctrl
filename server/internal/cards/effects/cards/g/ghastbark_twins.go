package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghastbark Twins — Creature — Treefolk {5}{G}{G}, 7/7
// (Core Set 2019, #5527):
//
//	"Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
//	This creature can block an additional creature each combat."
//
// Two-Headed Giant of Foriys's text: trample, and a second block
// (CR 509.1a) through CanBlockAdditional.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bdd0d01d-b9f2-4c3f-9b3b-04e122013de5",
		Name:            "Ghastbark Twins",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static:          []game.StaticAbility{effects.CanBlockAdditional(effects.SelfOnly, 1)},
	})
}
