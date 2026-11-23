package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barrage of Expendables — Enchantment {R} (Magic 2014, #5563):
//
//	"{R}, Sacrifice a creature: This enchantment deals 1 damage to any target."
//
// Goblin Bombardment at {R} a shot: the enchantment is the damage's source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da7600f6-5d4c-46bf-936a-9e0f0d508e19",
		Name:         "Barrage of Expendables",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Sacrifice a creature: This enchantment deals 1 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.SacrificeACreature()),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.DamageToFirstTarget(1),
		}},
	})
}
