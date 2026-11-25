package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Transguild Promenade — Land (Commander Legends):
//
//	"This land enters tapped.
//	 When this land enters, sacrifice it unless you pay {1}.
//	 {T}: Add one mana of any color."
//
// Rupture Spire's text exactly, so the same shared spec (EntersTappedSacrificeUnlessPaySpec).
//
// No simplification.
func init() {
	effects.Register(effects.EntersTappedSacrificeUnlessPaySpec("98334bfa-c516-4c20-bdc5-9e32e7127adc", "Transguild Promenade"))
}
