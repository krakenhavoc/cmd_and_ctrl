package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aven Archer — Creature — Bird Soldier Archer {3}{W}{W}, 2/2 (Odyssey,
// #5613):
//
//	"Flying
//	 {2}{W}, {T}: This creature deals 2 damage to target attacking or
//	 blocking creature."
//
// Catapult Squad's damage row behind a mana-and-tap cost: the target
// must be attacking or blocking at announce and as the ability resolves
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "914ce8f6-5e92-436d-8541-1c72da8577dc",
		Name:            "Aven Archer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}, {T}: This creature deals 2 damage to target attacking or blocking creature.",
			Cost:    effects.Plus(effects.ManaCost("{2}{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
			Effect:  effects.DealDamageToTheTarget(2),
		}},
	})
}
