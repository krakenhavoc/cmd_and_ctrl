package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "ec0ea7f7-52ce-40d1-b34c-e36dd4b26120",
		Name:         "Keldon Megaliths",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:     "Hellbent — {1}{R}, {T}: This land deals 1 damage to any target. Activate only if you have no cards in hand.",
			Purpose:   effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:      effects.Plus(effects.ManaCost("{1}{R}"), effects.TapCost()),
			Condition: effects.NoCardsInHand(),
			Targets:   effects.TargetAny(),
			Effect:    effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
