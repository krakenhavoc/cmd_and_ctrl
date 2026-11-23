package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silent Dart — Artifact {1} (Guilds of Ravnica, #5525):
//
//	"{4}, {T}, Sacrifice this artifact: It deals 3 damage to target creature."
//
// Bear Trap's ability at {4}: the tap and the sacrifice are paid at
// announcement, so the Dart is gone before the damage resolves and the
// damage comes from its last-known information (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "912bc348-9ba3-4001-90a8-104d2bb17527",
		Name:         "Silent Dart",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}, Sacrifice this artifact: It deals 3 damage to target creature.",
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(3),
		}},
	})
}
