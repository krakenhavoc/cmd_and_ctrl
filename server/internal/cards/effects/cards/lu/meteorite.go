package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "0c3c3bcc-d485-4a01-92fe-8bf29c0fc926",
		Name:         "Meteorite",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Meteorite — 2 damage to any target", effects.SourceDealsDamageToEachLegalTarget(2)), effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
	})
}
