package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nebula Dragon — Creature — Dragon {6}{R}, 4/4 (Edge of Eternities,
// #4461):
//
//	"Flying
//	 When this creature enters, it deals 3 damage to any target."
//
// Ghitu Slinger's entry trigger with 3 in place of 2: the Dragon is the
// source of the damage, and a target gone by resolution takes none
// (sourceDealsDamageToEachLegalTarget, CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "724b4a59-9c69-4f2b-a436-72183c62f1ff",
		Name:            "Nebula Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Nebula Dragon — 3 damage to any target", sourceDealsDamageToEachLegalTarget(3)), TargetAny()),
				ForTargets(DamageToTarget(0, 3))),
		},
	})
}
