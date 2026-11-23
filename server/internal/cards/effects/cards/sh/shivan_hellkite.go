package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shivan Hellkite — {5}{R}{R} Creature — Dragon 5/5:
//
//	"Flying
//	 {1}{R}: This creature deals 1 damage to any target."
//
// No tap in the cost, so it can ping as often as its controller can
// pay, summoning sick or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "069be6a0-9545-4fdc-ad40-957fe307db7a",
		Name:            "Shivan Hellkite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature deals 1 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.ManaCost("{1}{R}"),
			Targets: effects.TargetAny(),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
