package cardse

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Elfhame Palace — Land (Starter Commander Decks):
//
//	"This land enters tapped.
//	 {T}: Add {G} or {W}."
//
// Cinder Barrens' shape: it enters tapped (a replacement, CR 614.1c) and
// taps for either colour.
//
// No simplification.
func init() {
	effects.R42RegisterTappedDual("cade8b94-2998-4d23-87bb-9fbdddd19dea", "Elfhame Palace", "G", "W")
}
