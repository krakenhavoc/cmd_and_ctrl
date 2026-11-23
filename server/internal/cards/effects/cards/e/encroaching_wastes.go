package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Encroaching Wastes — Land (Magic 2014, #5563):
//
//	"{T}: Add {C}.
//	 {4}, {T}, Sacrifice this land: Destroy target nonbasic land."
//
// Wasteland at {4} more: a colorless land whose sacrifice destroys a nonbasic
// land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43144f06-079b-4515-a03a-01ea3e90d586",
		Name:         "Encroaching Wastes",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}, Sacrifice this land: Destroy target nonbasic land.",
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target nonbasic land", effects.Land(), effects.B10NonbasicLand()),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
