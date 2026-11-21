package cardsu

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Urborg Volcano — Land (Starter Commander Decks):
//
//	"This land enters tapped.
//	 {T}: Add {B} or {R}."
//
// Cinder Barrens' shape: it enters tapped (a replacement, CR 614.1c) and
// taps for either colour.
//
// No simplification.
func init() {
	effects.R42RegisterTappedDual("018129dd-f578-4fc9-b6f8-b13ed0c75a8d", "Urborg Volcano", "B", "R")
}
