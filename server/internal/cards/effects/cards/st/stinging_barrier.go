package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stinging Barrier — Creature — Wall {2}{U}{U}, 0/4 (Mercadian Masques, S82
// OLD, #5619):
//
//	"Defender
//	 {U}, {T}: This creature deals 1 damage to any target."
//
// Defender is printed data. The {T} is the Wall's own, so it is summoning-sick
// the turn it arrives (CR 302.6); the damage is its own, to any target, checked
// again as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b56f375f-35b3-4d96-9749-c059c2d29c8b",
		Name:            "Stinging Barrier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: This creature deals 1 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
