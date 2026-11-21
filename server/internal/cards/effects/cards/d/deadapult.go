package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deadapult — Enchantment {2}{R} (Dominaria Remastered, #5948):
//
//	"{R}, Sacrifice a Zombie: This enchantment deals 2 damage to any
//	 target."
//
// The Zombie is a cost (any Zombie its controller controls, a changeling
// included, read through the layers), and the damage is the enchantment's,
// not the Zombie's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e0fea29-0fd5-4535-b1df-cd66e50662cc",
		Name:         "Deadapult",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Sacrifice a Zombie: This enchantment deals 2 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Cost:    effects.Plus(effects.ManaCost("{R}"), game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Zombie", effects.Subtype("Zombie"))}),
			Targets: effects.TargetAny(),
			Effect:  effects.DealDamageToTheTarget(2),
		}},
	})
}
