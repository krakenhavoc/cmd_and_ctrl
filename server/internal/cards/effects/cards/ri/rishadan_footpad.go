package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rishadan Footpad — Creature — Human Pirate {3}{U}, 2/2 (Mercadian Masques, #5619):
//
//	"When this creature enters, each opponent sacrifices a permanent of
//	 their choice unless they pay {2}."
//
// Each opponent is asked for {2} as the trigger resolves; one who does not
// pay sacrifices a permanent of their choice. Shared body:
// RishadanEachOpponentSacrificesUnlessTheyPay (s82_old_b696_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a59f147b-3e03-4e82-aa34-48c220ea32e6",
		Name:         "Rishadan Footpad",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rishadan Footpad — each opponent sacrifices a permanent unless they pay {2}",
				effects.RishadanEachOpponentSacrificesUnlessTheyPay("Rishadan Footpad", "{2}")),
		},
	})
}
