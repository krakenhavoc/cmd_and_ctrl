package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gateway Plaza — Land — Gate (War of the Spark, #5520):
//
//	"This land enters tapped.
//	 When this land enters, sacrifice it unless you pay {1}.
//	 {T}: Add one mana of any color."
//
// Rupture Spire's text word for word: EntersTappedSacrificeUnlessPaySpec asks
// for the {1} as the trigger resolves and only while the land is still the
// object that entered (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.EntersTappedSacrificeUnlessPaySpec("a6543f71-0326-4e1f-b58f-9ce325d5d036", "Gateway Plaza"))
}
