package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Venerable Knight — Creature — Human Knight {W}, 2/1 (Throne of Eldraine, #5516):
//
//	"When this creature dies, put a +1/+1 counter on target Knight you control."
//
// A targeted dies trigger; with no Knight left it is removed (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ef027846-be81-4959-a6b5-56bd01b1e68a",
		Name:         "Venerable Knight",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisDies("Venerable Knight — put a +1/+1 counter on target Knight you control", rfCreatureFCountersOnTarget(1)),
				TargetCreature("target Knight you control", HasSubtype("Knight"), YouControl())),
		},
	})
}
