package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elite Archers — Creature — Human Soldier Archer {5}{W}, 3/3
// (Urza's Saga):
//
//	"{T}: This creature deals 3 damage to target attacking or blocking
//	 creature."
//
// Heavy Ballista's row for 3.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2230ec18-4dcf-435e-b495-1905cb94e8b8",
		Name:         "Elite Archers",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 3 damage to target attacking or blocking creature.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(3),
		}},
	})
}
