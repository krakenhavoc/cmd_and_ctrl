package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hickory Woodlot — Land (Mercadian Masques, #5619):
//
//	"This land enters tapped with two depletion counters on it.
//	 {T}, Remove a depletion counter from this land: Add {G}{G}. If there are no depletion counters on this land, sacrifice it."
//
// Built by MmqDepletionLand (s82_old_b693_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.MmqDepletionLand("26259c65-8f4e-42a6-b8a7-f65c36d35c4d", "Hickory Woodlot", "G"))
}
