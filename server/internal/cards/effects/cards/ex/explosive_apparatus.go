package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Explosive Apparatus — Artifact {1} (Shadows over Innistrad, #5543):
//
//	"{3}, {T}, Sacrifice this artifact: It deals 2 damage to any target."
//
// Aeolipile's sacrifice-for-two-damage at {3}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab8301da-f9f7-4d90-afae-bd7e60834c22",
		Name:         "Explosive Apparatus",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice this artifact: It deals 2 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetAny(),
			Effect:  effects.DealDamageToTheTarget(2),
		}},
	})
}
