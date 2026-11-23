package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prodigal Pyromancer — Creature — Human Wizard {2}{R}, 1/1 (Planar Chaos, #5595):
//
//	"{T}: This creature deals 1 damage to any target."
//
// Anaba Shaman's ping without the mana: the {T} keeps it off the turn it
// arrives (CR 302.6), and a target that left in response is not damaged
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29b3cb3e-aeb1-43eb-8095-9f7881aa34bd",
		Name:         "Prodigal Pyromancer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
