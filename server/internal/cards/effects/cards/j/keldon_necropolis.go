package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keldon Necropolis — Legendary Land (Invasion, S82 older sets batch 663, #5616):
//
//	"{T}: Add {C}.
//	 {4}{R}, {T}, Sacrifice a creature: Keldon Necropolis deals 2 damage to
//	 any target."
//
// A colorless mana ability and Blasting Station's ping: the sacrifice (any
// creature, this land is not one) and the {4}{R} are paid as the ability is
// activated, so the sacrificed creature's death triggers go on the stack above
// it, and the land deals the damage to what was announced if that is still
// legal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea4d6fcd-21e0-4e9f-b406-a89042998d98",
		Name:         "Keldon Necropolis",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{R}, {T}, Sacrifice a creature: Keldon Necropolis deals 2 damage to any target",
			Cost:    effects.Plus(effects.ManaCost("{4}{R}"), effects.TapCost(), effects.SacrificeACreature()),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.B33DamageChosenTargetFromSource(2),
		}},
	})
}
