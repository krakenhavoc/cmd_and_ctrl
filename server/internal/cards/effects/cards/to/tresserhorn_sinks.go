package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tresserhorn Sinks — Snow Land (Coldsnap, #5597):
//
//	"This land enters tapped.
//	 {T}: Add {B} or {R}."
//
// The tapped entry is the CR 614 self-replacement and the mana ability the
// printed one; the snow supertype is type-line data. No basic land types, so
// the two-colour ability is declared (Dominaria United's duals).
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB566SnowTapland("74a6597c-13ea-47bd-adb3-45671897b11b", "Tresserhorn Sinks", "B", "R"))
}
