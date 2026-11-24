package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raise the Draugr — Instant {1}{B} (Kaldheim):
//
//	"Choose one —
//	 • Return target creature card from your graveyard to your hand.
//	 • Return two target creature cards that share a creature type from your
//	   graveyard to your hand."
//
// Return from Extinction's text at instant speed: the second bullet's second
// clause is SharesACreatureTypeWithTargetCard over the first, and a changeling
// card shares a type with any creature card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8727284f-e9a6-49ea-9fb8-1076883ceecf",
		Name:         "Raise the Draugr",
		Completeness: effects.CompletenessFull,
		Modes:        effects.ReturnOneOrTwoSharingCreatureCardsModes(),
	})
}
