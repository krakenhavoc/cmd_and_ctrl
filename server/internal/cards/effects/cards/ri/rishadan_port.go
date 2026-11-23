package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rishadan Port — Land:
//
//	"{T}: Add {C}.
//	 {1}, {T}: Tap target land."
//
// The second ability is the shared tapChosenPermanent body; the target must
// still be a land as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f3e8dc56-2810-474e-a6a7-9c3555f94ae9",
		Name:         "Rishadan Port",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Tap target land.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
