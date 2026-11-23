package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soldier Replica — Artifact Creature — Soldier {3}, 1/3 (Mirrodin, #5606):
//
//	"{1}{W}, Sacrifice this creature: It deals 3 damage to target
//	 attacking or blocking creature."
//
// Divebomber Griffin's row paid with {1}{W} instead of a tap: the
// Replica is sacrificed as the cost and deals the damage as it last
// existed (CR 608.2h); a target no longer attacking or blocking as
// the ability resolves is not dealt it (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "702693f0-09f6-4f19-8dc9-6b31641e3347",
		Name:         "Soldier Replica",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, Sacrifice this creature: It deals 3 damage to target attacking or blocking creature.",
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
			Effect:  effects.B33DamageChosenTargetFromSource(3),
		}},
	})
}
