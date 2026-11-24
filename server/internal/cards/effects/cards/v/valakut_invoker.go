package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Valakut Invoker — Creature — Human Shaman {2}{R}, 2/3 (Commander Legends):
//
//	"{8}: This creature deals 3 damage to any target."
//
// A plain mana-cost activation with no tap, so summoning sickness does not
// apply (CR 302.6). The damage is dealt by the Invoker, to a target
// re-checked on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "17ed18cf-33e5-4861-8a36-8f7bb91971f9",
		Name:         "Valakut Invoker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{8}: This creature deals 3 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
			Cost:    effects.ManaCost("{8}"),
			Targets: effects.TargetAny(),
			Effect:  effects.DealDamageToTheTarget(3),
		}},
	})
}
