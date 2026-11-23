package cardsj

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Jwar Isle Refuge — Land (Starter Commander Decks):
//
//	"This land enters tapped.
//	 When this land enters, you gain 1 life.
//	 {T}: Add {U} or {B}."
//
// The gain land of the Khans cycle: the shared RegisterGainLand (enters
// tapped, the life as a real enters trigger, the two-colour pipe).
//
// No simplification.
func init() {
	effects.RegisterGainLand("8b96f837-7c32-473a-b5ae-1d66527eaf7b", "Jwar Isle Refuge", "U", "B")
}
