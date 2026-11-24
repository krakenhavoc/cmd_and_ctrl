package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rupture Spire — Land (Conflux, #5587):
//
//	"This land enters tapped.
//	 When this land enters, sacrifice it unless you pay {1}.
//	 {T}: Add one mana of any color."
//
// Command Bridge's land with a mana payment in place of the tap: the
// enters-tapped replacement (CR 614.1c), then a trigger that asks for
// {1} and sacrifices the land on a decline, through the shared
// upkeepSacrificeSource branch. A Spire that left in response, or left
// and came back (CR 400.7), asks nothing.
//
// Archway Commons (Strixhaven: School of Mages, #5502) prints the same text, so
// both are effects.EntersTappedSacrificeUnlessPaySpec.
//
// No simplification.
func init() {
	effects.Register(effects.EntersTappedSacrificeUnlessPaySpec("7eadffcb-1e15-44c1-b1db-78c71b8ec1ce", "Rupture Spire"))
}
