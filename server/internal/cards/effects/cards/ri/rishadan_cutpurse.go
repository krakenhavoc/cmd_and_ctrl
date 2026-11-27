package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rishadan Cutpurse — Creature — Human Pirate {2}{U}, 1/1 (Mercadian Masques, #5619):
//
//	"When this creature enters, each opponent sacrifices a permanent of
//	 their choice unless they pay {1}."
//
// Each opponent is asked for {1} as the trigger resolves; one who does not
// pay sacrifices a permanent of their choice. Shared body:
// RishadanEachOpponentSacrificesUnlessTheyPay (s82_old_b696_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "147e0e74-f76b-4f0a-951f-d39afdd72c02",
		Name:         "Rishadan Cutpurse",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rishadan Cutpurse — each opponent sacrifices a permanent unless they pay {1}",
				effects.RishadanEachOpponentSacrificesUnlessTheyPay("Rishadan Cutpurse", "{1}")),
		},
	})
}
