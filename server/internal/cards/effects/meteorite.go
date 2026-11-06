package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Meteorite — Artifact {5} (Avatar: The Last Airbender Eternal, #5075):
//
//	"When this artifact enters, it deals 2 damage to any target.
//	 {T}: Add one mana of any color."
//
// Ghitu Slinger's enters trigger on a mana rock. A target that left
// before the trigger resolves takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c3c3bcc-d485-4a01-92fe-8bf29c0fc926",
		Name:         "Meteorite",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Meteorite — 2 damage to any target", sourceDealsDamageToEachLegalTarget(2)), TargetAny()), ForTargets(DamageToTarget(0, 2))),
		},
	})
}
