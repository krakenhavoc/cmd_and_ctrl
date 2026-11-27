package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sandstone Needle — Land (Mercadian Masques, #5619):
//
//	"This land enters tapped with two depletion counters on it.
//	 {T}, Remove a depletion counter from this land: Add {R}{R}. If there are no depletion counters on this land, sacrifice it."
//
// Built by MmqDepletionLand (s82_old_b693_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.MmqDepletionLand("c8e0a1a5-8188-4677-9d8a-a18eb593343a", "Sandstone Needle", "R"))
}
