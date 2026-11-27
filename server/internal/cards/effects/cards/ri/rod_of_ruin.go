package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rod of Ruin — Artifact {4} (Magic 2014, S82 old sets):
//
//	"{3}, {T}: This artifact deals 1 damage to any target."
//
// Barrage of Expendables' ping with a mana-and-tap cost: the damage is the
// artifact's own, dealt to the target if it is still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c29a04ab-4e14-45d8-a993-3fffd0e2f6eb",
		Name:         "Rod of Ruin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: This artifact deals 1 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.DamageToFirstTarget(1),
		}},
	})
}
