package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// War-Torch Goblin — Creature — Goblin Warrior {R}, 1/1 (Ravnica: City of Guilds):
//
//	"{R}, Sacrifice this creature: It deals 2 damage to target blocking creature."
//
// Bloodpyre Elemental's shape with a blocking-creature clause. The Goblin is
// sacrificed as a cost, so the damage is dealt by its last-known information
// (CR 608.2h); the target is chosen before the sacrifice and checked again as
// the ability resolves (CR 608.2b), so a creature that stopped blocking in
// response is not hit.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "58d39991-420b-4eca-967d-49ad3f9a4702",
		Name:         "War-Torch Goblin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Sacrifice this creature: It deals 2 damage to target blocking creature.",
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target blocking creature", effects.BlockingCreature()),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.DealDamageToTheTarget(2),
		}},
	})
}
