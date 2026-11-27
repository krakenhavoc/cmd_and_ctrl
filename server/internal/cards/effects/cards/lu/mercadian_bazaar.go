package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mercadian Bazaar — Land (Mercadian Masques, #5619):
//
//	"This land enters tapped.
//	 {T}: Put a storage counter on this land.
//	 {T}, Remove any number of storage counters from this land: Add {R} for each storage counter removed this way."
//
// Built by MmqStorageLand (s82_old_b693_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.MmqStorageLand("b02ab3c7-fe4a-443c-b860-ba971d3301b0", "Mercadian Bazaar", "R"))
}
