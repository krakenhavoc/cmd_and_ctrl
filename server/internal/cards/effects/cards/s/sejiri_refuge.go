package cardss

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Sejiri Refuge — Land (Starter Commander Decks):
//
//	"This land enters tapped.
//	 When this land enters, you gain 1 life.
//	 {T}: Add {W} or {U}."
//
// The gain land of the Khans cycle: the shared RegisterGainLand (enters
// tapped, the life as a real enters trigger, the two-colour pipe).
//
// No simplification.
func init() {
	effects.RegisterGainLand("73b3e242-075d-4c4d-9b09-6fef1633c348", "Sejiri Refuge", "W", "U")
}
