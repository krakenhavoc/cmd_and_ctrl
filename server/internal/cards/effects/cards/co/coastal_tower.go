package cardsco

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Coastal Tower — Land (Starter Commander Decks):
//
//	"This land enters tapped.
//	 {T}: Add {W} or {U}."
//
// Cinder Barrens' shape: it enters tapped (a replacement, CR 614.1c) and
// taps for either colour.
//
// No simplification.
func init() {
	effects.GapB183RegisterTappedDual("f0367926-4380-4d22-8cb8-46c2076f102a", "Coastal Tower", "W", "U")
}
