package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rakdos Ickspitter — Creature — Thrull {1}{B}{R}, 1/1 (Dissension, #5598):
//
//	"{T}: This creature deals 1 damage to target creature and that
//	 creature's controller loses 1 life."
//
// disIckspitterEffect: the damage, then the life loss for the target's
// controller, both only while the target is still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "329f6e2e-6c37-4e98-bd45-bd252f1a2bd9",
		Name:         "Rakdos Ickspitter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to target creature and that creature's controller loses 1 life.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.DisIckspitterEffect,
		}},
	})
}
