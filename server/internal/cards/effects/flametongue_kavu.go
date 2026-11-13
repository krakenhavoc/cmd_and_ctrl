package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flametongue Kavu — Creature — Kavu {3}{R}, 4/2 (Planeshift, #5615):
//
//	"When this creature enters, it deals 4 damage to target creature."
//
// Firemaw Kavu's enters trigger at 4. The target is mandatory and may be
// any creature, the Kavu included: with no other creature on the
// battlefield it must target itself (CR 603.3d), as in paper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b215196a-3716-451f-bdbd-d16c05ace5b4",
		Name:         "Flametongue Kavu",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Flametongue Kavu — 4 damage to target creature", sourceDealsDamageToEachLegalTarget(4)),
				TargetCreature("target creature")), ForTargets(DamageToTarget(0, 4))),
		},
	})
}
