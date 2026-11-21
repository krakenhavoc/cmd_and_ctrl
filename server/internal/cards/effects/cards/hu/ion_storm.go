package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ion Storm — Enchantment {2}{R} (Magic Commander Collection, #5908):
//
//	"{1}{R}, Remove a +1/+1 counter or a charge counter from a permanent
//	 you control: This enchantment deals 2 damage to any target."
//
// The counter cost is an any-kind removal narrowed to two kinds
// (RemoveCountersOfKindsFrom): the activator picks the permanent and which of
// the two kinds comes off it, and nothing else pays. It is a cost (CR 118.3),
// so no replacement or counter-doubler has a say and the permanent is chosen,
// not targeted. The damage is Ion Storm's own, dealt by the enchantment.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "448e0cf2-c97c-4dc0-96ef-91f1eb75d97c",
		Name:         "Ion Storm",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, Remove a +1/+1 counter or a charge counter from a permanent you control: This enchantment deals 2 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Cost: effects.Plus(effects.ManaCost("{1}{R}"),
				effects.RemoveCountersOfKindsFrom(1, "a permanent you control", []string{game.CounterPlusOne, "charge"})),
			Targets: effects.TargetAny(),
			Effect:  effects.DealDamageToTheTarget(2),
		}},
	})
}
