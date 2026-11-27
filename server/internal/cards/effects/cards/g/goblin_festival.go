package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Festival — {1}{R} Enchantment (Urza's Destiny, S82 OLD batch 704, #5621):
//
//	"{2}: This enchantment deals 1 damage to any target. Flip a coin. If you
//	 lose the flip, choose one of your opponents. That player gains control of
//	 this enchantment."
//
// The damage, then the call; a lost flip asks the controller for one of their
// opponents, who gains control of the enchantment indefinitely if it is still
// the same permanent (S82OldB704GoblinFestival).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "de21b90f-a268-46d6-b977-d5b0856176f3",
		Name:         "Goblin Festival",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: This enchantment deals 1 damage to any target. Flip a coin. If you lose the flip, choose one of your opponents. That player gains control of this enchantment.",
			Cost:    effects.ManaCost("{2}"),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.S82OldB704GoblinFestival,
		}},
	})
}
