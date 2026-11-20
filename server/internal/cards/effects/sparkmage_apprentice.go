package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sparkmage Apprentice — Creature — Human Wizard {1}{R}, 1/1 (Ravnica: City
// of Guilds, #5600):
//
//	"When this creature enters, it deals 1 damage to any target."
//
// Skeleton Archer's targeted enters trigger: the creature is the source of
// the 1 damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69003cbc-8e5a-41dc-ac53-18f6e776217b",
		Name:         "Sparkmage Apprentice",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Sparkmage Apprentice — 1 damage to any target",
				sourceDealsDamageToEachLegalTarget(1)), TargetAny()), ForTargets(DamageToTarget(0, 1))),
		},
	})
}
