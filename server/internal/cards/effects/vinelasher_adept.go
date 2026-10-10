package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "b5918fb0-c33f-4e33-ac6d-8f94723d4296",
		Name:            "Vinelasher Adept",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{Targeting(
			WhenThisEnters("Vinelasher Adept — put three +1/+1 counters on target creature",
				rfCreatureFCountersOnTarget(3)),
			TargetCreature("target creature"),
		)},
		Activated: []ActivatedAbility{BasicLandcycling("{2}")},
	})
}
