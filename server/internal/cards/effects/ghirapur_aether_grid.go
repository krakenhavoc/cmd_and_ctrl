package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "05d75b10-4f20-4083-af27-f85c6e207fd0",
		Name:         "Ghirapur Aether Grid",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "Tap two untapped artifacts you control: This enchantment deals 1 damage to any target.",
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  2,
				Filter: TargetPermanent("two untapped artifacts you control", Artifact()),
				Label:  "two untapped artifacts you control",
			}},
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  sourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
