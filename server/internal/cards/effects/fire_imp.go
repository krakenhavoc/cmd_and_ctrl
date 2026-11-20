package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fire Imp — Creature — Imp {2}{R}, 2/1 (Portal, #5630):
//
//	"When this creature enters, it deals 2 damage to target creature."
//
// Flametongue Kavu's enters trigger, with 2 damage. With no creature
// to target, the trigger is removed from the stack (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5bd806e7-3f9b-4bb4-9708-f87a578f531e",
		Name:         "Fire Imp",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Fire Imp — 2 damage to target creature", sourceDealsDamageToEachLegalTarget(2)),
				TargetCreature("target creature")), ForTargets(DamageToTarget(0, 2))),
		},
	})
}
