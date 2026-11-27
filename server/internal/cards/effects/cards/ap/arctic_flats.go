package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arctic Flats — Snow Land (Coldsnap, #5597):
//
//	"This land enters tapped.
//	 {T}: Add {G} or {W}."
//
// The tapped entry is the CR 614 self-replacement and the mana ability the
// printed one; the snow supertype is type-line data. No basic land types, so
// the two-colour ability is declared (Dominaria United's duals).
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB566SnowTapland("841f0a1d-6f84-45fb-81cc-f2d99fa9e9a2", "Arctic Flats", "G", "W"))
}
