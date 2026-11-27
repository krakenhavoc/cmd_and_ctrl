package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rushwood Grove — Land (Mercadian Masques, #5619):
//
//	"This land enters tapped.
//	 {T}: Put a storage counter on this land.
//	 {T}, Remove any number of storage counters from this land: Add {G} for each storage counter removed this way."
//
// Built by MmqStorageLand (s82_old_b693_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.MmqStorageLand("ccb2f92e-69c0-415c-81cd-52c384b3b233", "Rushwood Grove", "G"))
}
