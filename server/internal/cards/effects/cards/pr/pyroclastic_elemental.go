package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyroclastic Elemental — Creature — Elemental {3}{R}{R}, 5/4 (Core Set 2020, #5518):
//
//	"{1}{R}{R}: This creature deals 1 damage to target player."
//
// A player-only ping with no {T}, so it can be activated any number of
// times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f37e3a6-ce3a-472f-ac45-0d9b23a31a11",
		Name:         "Pyroclastic Elemental",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}{R}: This creature deals 1 damage to target player.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.ManaCost("{1}{R}{R}"),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
