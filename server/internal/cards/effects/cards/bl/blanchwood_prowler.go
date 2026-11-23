package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blanchwood Prowler — Creature — Elemental {1}{G}, 1/1 (The Brothers' War):
//
//	"When this creature enters, mill three cards. You may put a land card from
//	 among the cards milled this way into your hand. If you don't, put a
//	 +1/+1 counter on this creature."
//
// Ainok Wayfarer's mill-then-take shape (bro6MillThreeTakeOne), offering
// a land card from the cards this mill put into the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8d1a367a-542e-4b94-b466-c922c5fb717b",
		Name:         "Blanchwood Prowler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Bro6MillThreeTakeOne("Blanchwood Prowler", "blanchwood-prowler", "a land card", effects.Land()),
		},
	})
}
