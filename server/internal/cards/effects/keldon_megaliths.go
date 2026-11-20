package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Keldon Megaliths — Land (Future Sight, #5594):
//
//	"This land enters tapped.
//	 {T}: Add {R}.
//	 Hellbent — {1}{R}, {T}: This land deals 1 damage to any target.
//	 Activate only if you have no cards in hand."
//
// A tapped red land with Sea Gate Wreckage's hellbent gate (NoCardsInHand)
// on a ping. Entering tapped is the CR 614 self-replacement, as on the
// Temples.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ec0ea7f7-52ce-40d1-b34c-e36dd4b26120",
		Name:         "Keldon Megaliths",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Activated: []ActivatedAbility{{
			Label:     "Hellbent — {1}{R}, {T}: This land deals 1 damage to any target. Activate only if you have no cards in hand.",
			Purpose:   ForTargets(DamageToTarget(0, 1)),
			Cost:      Plus(ManaCost("{1}{R}"), TapCost()),
			Condition: NoCardsInHand(),
			Targets:   TargetAny(),
			Effect:    sourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
