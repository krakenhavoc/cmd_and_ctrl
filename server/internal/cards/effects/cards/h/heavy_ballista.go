package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heavy Ballista — Creature — Human Soldier {3}{W}, 2/3 (Weatherlight, #5629):
//
//	"{T}: This creature deals 2 damage to target attacking or blocking
//	 creature."
//
// A tap ability with a combat target: a creature that is no longer
// attacking or blocking as it resolves is an illegal target (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b4219836-c802-49e2-9a66-f03ca8f13895",
		Name:         "Heavy Ballista",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 2 damage to target attacking or blocking creature.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
