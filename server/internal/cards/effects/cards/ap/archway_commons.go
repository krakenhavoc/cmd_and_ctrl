package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Archway Commons — Land (Strixhaven: School of Mages, #5502):
//
//	"This land enters tapped.
//	 When this land enters, sacrifice it unless you pay {1}.
//	 {T}: Add one mana of any color."
//
// Rupture Spire's text word for word, so both are EntersTappedSacrificeUnlessPaySpec
// (s82_old_b051_helpers.go): the replacement that enters it tapped, the
// pay-unless as the trigger resolves, and the any-color mana ability.
//
// No simplification.
func init() {
	effects.Register(effects.EntersTappedSacrificeUnlessPaySpec("69c63055-ed44-4b32-b591-f3c6c2f3e7d1", "Archway Commons"))
}
