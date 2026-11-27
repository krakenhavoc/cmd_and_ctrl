package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Remote Farm — Land (Mercadian Masques, #5619):
//
//	"This land enters tapped with two depletion counters on it.
//	 {T}, Remove a depletion counter from this land: Add {W}{W}. If there are no depletion counters on this land, sacrifice it."
//
// Built by MmqDepletionLand (s82_old_b693_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.MmqDepletionLand("2c38f4c7-1b3f-42b4-a175-edab7acd6cc6", "Remote Farm", "W"))
}
