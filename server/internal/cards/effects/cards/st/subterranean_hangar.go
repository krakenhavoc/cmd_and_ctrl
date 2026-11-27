package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Subterranean Hangar — Land (Mercadian Masques, #5619):
//
//	"This land enters tapped.
//	 {T}: Put a storage counter on this land.
//	 {T}, Remove any number of storage counters from this land: Add {B} for each storage counter removed this way."
//
// Built by MmqStorageLand (s82_old_b693_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.MmqStorageLand("0bbd5a04-c281-4afb-98a1-657b4eca102c", "Subterranean Hangar", "B"))
}
