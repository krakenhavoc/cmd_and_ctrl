package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fountain of Cho — Land (Mercadian Masques, #5619):
//
//	"This land enters tapped.
//	 {T}: Put a storage counter on this land.
//	 {T}, Remove any number of storage counters from this land: Add {W} for each storage counter removed this way."
//
// Built by MmqStorageLand (s82_old_b693_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.MmqStorageLand("0799df10-b489-4f79-bf98-7a0c500b46a1", "Fountain of Cho", "W"))
}
