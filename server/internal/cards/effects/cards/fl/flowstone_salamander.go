package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flowstone Salamander — Creature — Salamander {3}{R}{R}, 3/4 (Tempest,
// S82 OLD batch 758, #5628):
//
//	"{R}: This creature deals 1 damage to target creature blocking it."
//
// The target is a creature blocking this one (BlockingSource, #1863), judged
// when the ability is activated and again as it resolves (CR 608.2b): a
// blocker removed from combat in response is no longer legal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a0e40eb1-a909-405b-8fb4-3759df8edf7f",
		Name:         "Flowstone Salamander",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature deals 1 damage to target creature blocking it.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.ManaCost("{R}"),
			Targets: effects.BlockingSource(effects.TargetCreature("target creature blocking it")),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
