package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blisterstick Shaman — Creature — Goblin Shaman {2}{R}, 2/1 (Mirrodin Besieged,
// #5576):
//
//	"When this creature enters, it deals 1 damage to any target."
//
// Skeleton Archer's enters ping: the Shaman is the damage source, and a
// target that has become illegal by resolution is not dealt damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f76fec2-6fd0-4d33-b39a-f1684922d478",
		Name:         "Blisterstick Shaman",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Blisterstick Shaman — 1 damage to any target",
				sourceDealsDamageToEachLegalTarget(1)), TargetAny()), ForTargets(DamageToTarget(0, 1))),
		},
	})
}
