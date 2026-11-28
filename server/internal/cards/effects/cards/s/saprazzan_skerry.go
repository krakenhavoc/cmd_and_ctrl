package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saprazzan Skerry — Land (Mercadian Masques, #5619):
//
//	"This land enters tapped with two depletion counters on it.
//	 {T}, Remove a depletion counter from this land: Add {U}{U}. If there are no depletion counters on this land, sacrifice it."
//
// Built by MmqDepletionLand (s82_old_b693_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.MmqDepletionLand("e4e6e796-39ce-4a63-8c61-c7c956d75d78", "Saprazzan Skerry", "U"))
}
