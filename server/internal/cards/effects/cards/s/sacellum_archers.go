package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sacellum Archers — Creature — Elf Archer {2}{G}, 2/3 (Conflux, #5587):
//
//	"{R}{W}, {T}: This creature deals 2 damage to target attacking or
//	 blocking creature."
//
// Aven Archer's row at {R}{W}. The {T} is held back by summoning
// sickness (CR 302.6), and a target no longer attacking or blocking when
// the ability resolves is not dealt damage (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b9117a1-cd2d-4478-beda-1049e0b0f8eb",
		Name:         "Sacellum Archers",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}{W}, {T}: This creature deals 2 damage to target attacking or blocking creature.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Cost:    effects.Plus(effects.ManaCost("{R}{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
			Effect:  effects.DealDamageToTheTarget(2),
		}},
	})
}
