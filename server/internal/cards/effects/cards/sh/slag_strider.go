package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slag Strider — Creature — Elemental {5}{R}{R}, 3/3 (Modern Horizons 2,
// #5501):
//
//	"Affinity for artifacts (This spell costs {1} less to cast for each
//	 artifact you control.)
//	 {1}, Sacrifice an artifact: This creature deals 1 damage to any
//	 target."
//
// Frogmite's affinity (CR 702.41a, generic mana only) and Orcish
// Mechanics' sacrifice-an-artifact ping with {1} for the {T}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2b154d55-37dd-4e5f-aa99-d4bc547cdec0",
		Name:         "Slag Strider",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for artifacts", effects.Artifact()),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice an artifact: This creature deals 1 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeN(1, "an artifact", effects.Artifact())),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
