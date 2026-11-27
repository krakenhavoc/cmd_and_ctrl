package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Opportunist — Creature — Human Soldier {2}{R}, 2/2 (Tempest, S82 OLD
// batch 757, #5628):
//
//	"{T}: This creature deals 1 damage to target creature that was dealt
//	 damage this turn."
//
// A tap ping whose target is restricted to creatures dealt damage this turn
// (Rooftop Assassin's DealtDamageToday filter), checked on activation and
// again on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b5524f7-3c68-4c7b-be1b-745ffcad8006",
		Name:         "Opportunist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to target creature that was dealt damage this turn.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature that was dealt damage this turn", effects.DealtDamageThisTurn()),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
