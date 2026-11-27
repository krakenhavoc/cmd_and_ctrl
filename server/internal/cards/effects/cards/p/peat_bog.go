package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Peat Bog — Land (Mercadian Masques, #5619):
//
//	"This land enters tapped with two depletion counters on it.
//	 {T}, Remove a depletion counter from this land: Add {B}{B}. If there are no depletion counters on this land, sacrifice it."
//
// Built by MmqDepletionLand (s82_old_b693_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.MmqDepletionLand("a176924c-78fc-4151-b2b5-1547b1114a40", "Peat Bog", "B"))
}
