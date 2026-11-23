package cardss

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Salt Marsh — Land (Starter Commander Decks):
//
//	"This land enters tapped.
//	 {T}: Add {U} or {B}."
//
// Cinder Barrens' shape: it enters tapped (a replacement, CR 614.1c) and
// taps for either colour.
//
// No simplification.
func init() {
	effects.R42RegisterTappedDual("6a40823f-130c-4416-a987-75e205d3e1dd", "Salt Marsh", "U", "B")
}
