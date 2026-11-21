package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mudbutton Torchrunner — Creature — Goblin Warrior {2}{R}, 1/1 (Jumpstart
// 2022, #5948):
//
//	"When this creature dies, it deals 3 damage to any target."
//
// Pitchburn Devils' dies trigger: the damage comes from the Torchrunner as
// it last existed on the battlefield (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "045b2b84-96bd-46fb-8aae-a33968e741c7",
		Name:         "Mudbutton Torchrunner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisDies("Mudbutton Torchrunner — 3 damage to any target", sourceDealsDamageToEachLegalTarget(3)),
				TargetAny()), ForTargets(DamageToTarget(0, 3))),
		},
	})
}
