package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dracosaur Auxiliary — Creature — Dinosaur Dragon Mount {4}{R}{R}:
//
//	"Flying, haste
//	 Whenever this creature attacks while saddled, it deals 2 damage to
//	 any target.
//	 Saddle 3"
//
// The damage is the Dracosaur's own (sourceDealsDamageToEachLegalTarget),
// so lifelink, deathtouch and damage prevention read the right source.
//
// No simplification.
func init() {
	attack := AttacksWhileSaddled("Dracosaur Auxiliary — 2 damage to any target", sourceDealsDamageToEachLegalTarget(2))
	attack.Targets = TargetAny()
	attack = TriggerWithPurpose(attack, ForTargets(DamageToTarget(0, 2)))
	Register(Spec{
		OracleID:        "457e3218-b2ef-4280-931a-5320987b9657",
		Name:            "Dracosaur Auxiliary",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Activated:       []ActivatedAbility{Saddle(3)},
		Triggered:       []game.TriggeredAbility{attack},
	})
}
