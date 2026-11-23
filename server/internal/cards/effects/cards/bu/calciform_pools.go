package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Calciform Pools — Land (Time Spiral, #5596):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Put a storage counter on this land.
//	 {1}, Remove X storage counters from this land: Add X mana in any
//	 combination of {W} and/or {U}."
//
// The white-blue storage land, built by tspB2StorageLand: the counter is a
// plain activated ability that uses the stack, and the payout is a mana
// ability that pays {1}, removes the X counters announced and adds X
// mana, each {W} or {U} (CR 605.1a).
//
// No simplification.
func init() {
	effects.Register(effects.TspB2StorageLand("2031bc31-81cc-407a-8615-29832f586bbc", "Calciform Pools", "W", "U"))
}
