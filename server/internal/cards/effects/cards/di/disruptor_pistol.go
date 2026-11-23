package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disruptor Pistol — Artifact {2} (Star Trek Commander, #4974):
//
//	"{4}, {T}, Sacrifice this artifact: It deals 5 damage to target
//	 creature."
//
// Exploding Barrel's damage ability without its discount or its timing
// limit, so it can be activated at instant speed. The Pistol is gone by
// the time the ability resolves; the damage is its, as it last existed
// (CR 608.2h), and a target that has left takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb8c2c0f-fc63-40c6-98c6-691e42b3b8c9",
		Name:         "Disruptor Pistol",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}, Sacrifice this artifact: It deals 5 damage to target creature.",
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 5)),
			Effect:  effects.DealDamageToTheTarget(5),
		}},
	})
}
