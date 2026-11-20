package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pitchburn Devils — Creature — Devil {4}{R}, 3/3 (Innistrad, #5572):
//
//	"When this creature dies, it deals 3 damage to any target."
//
// Pyre Spawn's death trigger: the target is chosen as the trigger goes
// on the stack, and the creature, now in the graveyard, is the source of
// the damage (CR 603.10a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "86e2bbdd-a788-4f76-9ebb-ef48d9bc9dcf",
		Name:         "Pitchburn Devils",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisDies("Pitchburn Devils — 3 damage to any target", sourceDealsDamageToEachLegalTarget(3)),
				TargetAny()), ForTargets(DamageToTarget(0, 3))),
		},
	})
}
