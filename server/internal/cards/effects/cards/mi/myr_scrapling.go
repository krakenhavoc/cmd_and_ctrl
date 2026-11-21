package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Myr Scrapling — Artifact Creature — Myr {1}, 1/1 (Modern Horizons 2,
// #5501):
//
//	"Sacrifice this creature: Put a +1/+1 counter on target creature."
//
// Shambling Shell's ability, with the shared PlusOneCounterOnTheTarget
// body (skipped if the target left, CR 608.2b): the sacrifice is the cost, so the Scrapling
// is gone before the counter is placed and can't take it itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a4ffc1e-2f5e-446a-b1e8-32385b3c083b",
		Name:         "Myr Scrapling",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Put a +1/+1 counter on target creature.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.PlusOneCounterOnTheTarget,
		}},
	})
}
