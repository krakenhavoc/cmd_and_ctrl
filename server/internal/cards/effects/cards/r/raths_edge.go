package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rath's Edge — Legendary Land (Nemesis, S82 OLD batch 680, #5618):
//
//	"{T}: Add {C}.
//	 {4}, {T}, Sacrifice a land: Rath's Edge deals 1 damage to any target."
//
// The land sacrificed is a cost (CR 602.2b) and may be Rath's Edge itself,
// which then deals the damage as it last existed (CR 608.2h). With no
// untapped Rath's Edge or no land to give up the ability can't be activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e10e84a7-d564-487a-ac64-5a001a45ee90",
		Name:         "Rath's Edge",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "{T}: Add {C}.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}, Sacrifice a land: Rath's Edge deals 1 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost(), effects.SacrificeN(1, "a land", effects.Land())),
			Targets: effects.TargetAny(),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
