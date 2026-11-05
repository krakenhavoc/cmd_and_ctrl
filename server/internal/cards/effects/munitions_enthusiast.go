package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Munitions Enthusiast — Creature — Klingon Warrior {1}{R}, 3/2 (Star
// Trek, #4973):
//
//	"Trample
//	 When this creature dies, it deals 1 damage to any target."
//
// The damage is dealt by the creature, which is in the graveyard by then:
// its last-known information is the source (CR 603.10a, Dynamite Diver's
// shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0013f005-2b49-48a4-8a74-9bbdadc88c9f",
		Name:            "Munitions Enthusiast",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisDies("Munitions Enthusiast — 1 damage to any target", sourceDealsDamageToEachLegalTarget(1)),
				TargetAny()), ForTargets(DamageToTarget(0, 1))),
		},
	})
}
