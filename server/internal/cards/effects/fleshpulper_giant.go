package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fleshpulper Giant — Creature — Giant {5}{R}{R}, 4/4 (Magic 2014, #5563):
//
//	"When this creature enters, you may destroy target creature with toughness 2 or less."
//
// An optional enters trigger (CR 603.5): the target is chosen as it goes on
// the stack, and the "you may" is asked as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7cdeecb5-354a-4a6c-989f-d8c9554befb7",
		Name:         "Fleshpulper Giant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(WhenThisEnters("Fleshpulper Giant — you may destroy target creature with toughness 2 or less", destroyFirstLegalCardTarget),
				TargetCreature("target creature with toughness 2 or less", ToughnessLE(2))),
				"Fleshpulper Giant — destroy the creature?"),
		},
	})
}
