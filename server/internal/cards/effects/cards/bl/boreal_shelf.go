package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boreal Shelf — Snow Land (Coldsnap, #5597):
//
//	"This land enters tapped.
//	 {T}: Add {W} or {U}."
//
// The tapped entry is the CR 614 self-replacement and the mana ability the
// printed one; the snow supertype is type-line data. No basic land types, so
// the two-colour ability is declared (Dominaria United's duals).
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB566SnowTapland("20b549e9-26db-40a5-8bcc-e8de6c559d32", "Boreal Shelf", "W", "U"))
}
