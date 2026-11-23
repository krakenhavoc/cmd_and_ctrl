package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scalding Cauldron — Artifact {1} (Throne of Eldraine, #5516):
//
//	"{3}, {T}, Sacrifice this artifact: It deals 3 damage to target creature."
//
// Bear Trap's ability without the flash. "It" is the sacrificed artifact,
// read from its last-known information.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dceb96a4-d3c5-43a2-93f6-50316e1e7a1b",
		Name:         "Scalding Cauldron",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice this artifact: It deals 3 damage to target creature.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(3),
		}},
	})
}
