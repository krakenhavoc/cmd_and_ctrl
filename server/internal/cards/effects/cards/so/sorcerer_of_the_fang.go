package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sorcerer of the Fang — Creature — Human Wizard Sorcerer {1}{B}, 1/3 (Core Set 2020, #5518):
//
//	"{5}{B}, {T}: This creature deals 2 damage to target opponent or
//	 planeswalker."
//
// Zealot of the God-Pharaoh's ability with a {T} in the cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c1d5a631-a963-43a4-a5c0-1ba23bd88d0f",
		Name:         "Sorcerer of the Fang",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{B}, {T}: This creature deals 2 damage to target opponent or planeswalker.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Cost:    effects.Plus(effects.ManaCost("{5}{B}"), effects.TapCost()),
			Targets: effects.TargetOpponentOrPlaneswalker(),
			Effect:  effects.DealDamageToTheTarget(2),
		}},
	})
}
