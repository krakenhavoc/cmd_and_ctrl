package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scaldkin — Creature — Elemental {3}{U}, 2/2 (Khans of Tarkir):
//
//	"Flying
//	 {2}{R}, Sacrifice this creature: It deals 2 damage to any target."
//
// Mogg Fanatic's sacrifice-for-damage at {2}{R} and 2: the sacrifice is a
// cost, so the damage is dealt by the Scaldkin as it last existed (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3fb2dbb3-e064-44bf-a8b6-10d58b1174d5",
		Name:            "Scaldkin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, Sacrifice this creature: It deals 2 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.SacrificeThis()),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.B33DamageChosenTargetFromSource(2),
		}},
	})
}
