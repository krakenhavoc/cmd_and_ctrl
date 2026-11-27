package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Highland Weald — Snow Land (Coldsnap, #5597):
//
//	"This land enters tapped.
//	 {T}: Add {R} or {G}."
//
// The tapped entry is the CR 614 self-replacement and the mana ability the
// printed one; the snow supertype is type-line data. No basic land types, so
// the two-colour ability is declared (Dominaria United's duals).
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB566SnowTapland("66d845a2-be4c-41bb-a357-3650461c03ed", "Highland Weald", "R", "G"))
}
