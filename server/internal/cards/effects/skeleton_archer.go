package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skeleton Archer — Creature — Skeleton Archer {3}{B}, 3/3
// (Foundations, #4465):
//
//	"When this creature enters, it deals 1 damage to any target."
//
// Mongoose Lizard's enters ping: the Archer is the damage source, and a
// target that has become illegal by resolution is not dealt damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3792763c-7aa0-4055-b8a1-3257e0443ef1",
		Name:         "Skeleton Archer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Skeleton Archer — 1 damage to any target",
				sourceDealsDamageToEachLegalTarget(1)), TargetAny()), ForTargets(DamageToTarget(0, 1))),
		},
	})
}
