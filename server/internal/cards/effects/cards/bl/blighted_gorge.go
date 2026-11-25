package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blighted Gorge — Land (Battle for Zendikar):
//
//	"{T}: Add {C}.
//	 {4}{R}, {T}, Sacrifice this land: It deals 2 damage to any target."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2cb0afd-781f-4cfa-b680-ed1edfa81868",
		Name:         "Blighted Gorge",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{R}, {T}, Sacrifice this land: It deals 2 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Cost:    effects.Plus(effects.ManaCost("{4}{R}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetAny(),
			Effect:  effects.DealDamageToTheTarget(2),
		}},
	})
}
