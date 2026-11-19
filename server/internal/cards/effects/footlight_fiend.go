package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Footlight Fiend — Creature — Devil {B/R}, 1/1 (Ravnica: Clue Edition):
//
//	"When this creature dies, it deals 1 damage to any target."
//
// Dynamite Diver's death trigger (damage by last-known information,
// CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77a5305a-2611-4a64-adb4-da131dcea233",
		Name:         "Footlight Fiend",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisDies("Footlight Fiend — 1 damage to any target", sourceDealsDamageToEachLegalTarget(1)),
				TargetAny()), ForTargets(DamageToTarget(0, 1))),
		},
	})
}
