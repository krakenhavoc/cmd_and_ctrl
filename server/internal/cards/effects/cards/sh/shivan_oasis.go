package cardssh

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Shivan Oasis — Land (Starter Commander Decks):
//
//	"This land enters tapped.
//	 {T}: Add {R} or {G}."
//
// Cinder Barrens' shape: it enters tapped (a replacement, CR 614.1c) and
// taps for either colour.
//
// No simplification.
func init() {
	effects.R42RegisterTappedDual("9b64c0f4-b917-4a0d-b033-30ab71aab807", "Shivan Oasis", "R", "G")
}
