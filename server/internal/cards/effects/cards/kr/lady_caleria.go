package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lady Caleria — Legendary Creature — Elf Archer {3}{G}{G}{W}{W}, 3/6
// (Legends, #5639):
//
//	"{T}: Lady Caleria deals 3 damage to target attacking or blocking
//	 creature."
//
// Tor Wauki's shape at 3: an AttackingOrBlocking target clause and the
// shared DealDamageToTheTarget body, dealt by Lady Caleria. The tap
// means summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "74284bbc-b011-4e16-b8d5-7ddacc77f9d7",
		Name:         "Lady Caleria",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Lady Caleria deals 3 damage to target attacking or blocking creature.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
			Effect:  effects.DealDamageToTheTarget(3),
		}},
	})
}
