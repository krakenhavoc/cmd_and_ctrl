package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tower of Calamities — Artifact {4} (Scars of Mirrodin, #5577):
//
//	"{8}, {T}: This artifact deals 12 damage to target creature."
//
// Aladdin's Ring's shape at 12 to a creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27618701-e8c9-4117-94f3-3836cf34990d",
		Name:         "Tower of Calamities",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{8}, {T}: This artifact deals 12 damage to target creature.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 12)),
			Cost:    effects.Plus(effects.ManaCost("{8}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.DealDamageToTheTarget(12),
		}},
	})
}
