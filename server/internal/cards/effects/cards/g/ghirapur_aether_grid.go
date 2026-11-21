package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghirapur Aether Grid — Enchantment {2}{R}:
//
//	"Tap two untapped artifacts you control: This enchantment deals 1
//	 damage to any target."
//
// Bladewheel Chariot's tap-others cost over artifacts, two of them. The
// cost has no {T} symbol, so an artifact creature that arrived this turn
// may pay it (CR 302.6); it does not target. The damage is the
// enchantment's own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "05d75b10-4f20-4083-af27-f85c6e207fd0",
		Name:         "Ghirapur Aether Grid",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "Tap two untapped artifacts you control: This enchantment deals 1 damage to any target.",
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  2,
				Filter: effects.TargetPermanent("two untapped artifacts you control", effects.Artifact()),
				Label:  "two untapped artifacts you control",
			}},
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
