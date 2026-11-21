package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vulshok Replica — Artifact Creature — Berserker {3}, 3/1 (Scars of Mirrodin, #5577):
//
//	"{1}{R}, Sacrifice this creature: It deals 3 damage to target player or
//	 planeswalker."
//
// Goblin Fireslinger's target clause (targetPlayerOrPlaneswalker) on
// Barbarian Lunatic's shape: the Replica is sacrificed as a cost and
// deals the damage as it last existed (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da8e2236-88ae-4f30-884b-b961231fba8f",
		Name:         "Vulshok Replica",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, Sacrifice this creature: It deals 3 damage to target player or planeswalker.",
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.SacrificeThis()),
			Targets: effects.TargetPlayerOrPlaneswalker(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
			Effect:  effects.DealDamageToTheTarget(3),
		}},
	})
}
