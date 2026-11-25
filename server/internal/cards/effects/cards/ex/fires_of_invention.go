package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fires of Invention — Enchantment {3}{R} (Throne of Eldraine, S82 old sets):
//
//	"You can cast spells only during your turn and you can cast no more than
//	 two spells each turn.
//	 You may cast spells with mana value less than or equal to the number of
//	 lands you control without paying their mana costs."
//
// Three statics. Both halves of the first are cast restrictions on its
// controller alone (YouCantCastUnless with the turn as the lifting condition,
// and a cap of two off the per-turn tally), so their presence on the battlefield
// is their duration. The second is Omnipresence's granted free offer with its
// cap following the lands its controller controls, reaching every zone; the
// offer is a cast, so the two-spell cap and the timing rules still apply to it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27e27681-9221-4e64-8c78-e153a86bc6e0",
		Name:         "Fires of Invention",
		Completeness: effects.CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			effects.YouCantCastUnless("You can cast spells only during your turn.", nil, effects.IsYourTurn),
			effects.OldB148YouCantCastMoreThan(2, "You can cast no more than two spells each turn."),
		},
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{effects.OldB148CastFreeUpToLandCount()},
	})
}
