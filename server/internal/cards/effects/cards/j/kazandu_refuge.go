package cardsj

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Kazandu Refuge — Land (Starter Commander Decks):
//
//	"This land enters tapped.
//	 When this land enters, you gain 1 life.
//	 {T}: Add {R} or {G}."
//
// The gain land of the Khans cycle: the shared RegisterGainLand (enters
// tapped, the life as a real enters trigger, the two-colour pipe).
//
// No simplification.
func init() {
	effects.RegisterGainLand("03332772-18b0-446a-a7e2-a4bdf38c4f0c", "Kazandu Refuge", "R", "G")
}
