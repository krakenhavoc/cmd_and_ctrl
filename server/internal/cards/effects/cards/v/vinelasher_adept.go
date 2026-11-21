package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vinelasher Adept — Creature — Rhino Soldier {4}{G}{G}, 2/4:
//
//	"Reach
//	 When this creature enters, put three +1/+1 counters on target
//	 creature.
//	 Basic landcycling {2}"
//
// The target is any creature, the Adept included; it is re-checked at
// resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b5918fb0-c33f-4e33-ac6d-8f94723d4296",
		Name:            "Vinelasher Adept",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.WhenThisEnters("Vinelasher Adept — put three +1/+1 counters on target creature",
				effects.RfCreatureFCountersOnTarget(3)),
			effects.TargetCreature("target creature"),
		)},
		Activated: []effects.ActivatedAbility{effects.BasicLandcycling("{2}")},
	})
}
