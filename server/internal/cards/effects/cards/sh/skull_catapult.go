package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skull Catapult — Artifact {4} (Ice Age, S82 OLD, #5636):
//
//	"{1}, {T}, Sacrifice a creature: This artifact deals 2 damage to any target."
//
// Barrage of Expendables's sacrifice-for-damage with a tap in the cost (so a Catapult
// can be used the turn it arrives: it is not a creature).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eef931d8-4048-4c33-bd8c-0f67d1083ee6",
		Name:         "Skull Catapult",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice a creature: This artifact deals 2 damage to any target.",
			Cost:    effects.Plus(effects.Plus(effects.ManaCost("{1}"), effects.TapCost()), effects.SacrificeACreature()),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.DamageToFirstTarget(2),
		}},
	})
}
