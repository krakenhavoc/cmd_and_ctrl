package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frost Marsh — Snow Land (Coldsnap, #5597):
//
//	"This land enters tapped.
//	 {T}: Add {U} or {B}."
//
// The tapped entry is the CR 614 self-replacement and the mana ability the
// printed one; the snow supertype is type-line data. No basic land types, so
// the two-colour ability is declared (Dominaria United's duals).
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB566SnowTapland("75a49bce-ab48-44d6-906d-6f51a30702ba", "Frost Marsh", "U", "B"))
}
