package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crackling Triton — Creature — Merfolk Wizard {2}{U}, 2/3 (Theros):
//
//	"{2}{R}, Sacrifice this creature: It deals 2 damage to any target."
//
// The sacrifice is a cost, so the damage is dealt by the Triton as it last
// existed (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8a327626-ce42-4ef8-8869-3468960c671e",
		Name:         "Crackling Triton",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, Sacrifice this creature: It deals 2 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.SacrificeThis()),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.B33DamageChosenTargetFromSource(2),
		}},
	})
}
