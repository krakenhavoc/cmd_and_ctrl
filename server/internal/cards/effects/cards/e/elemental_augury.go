package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elemental Augury — Enchantment {U}{B}{R} (Ice Age, S82 OLD #5636):
//
//	"{3}: Look at the top three cards of target player's library, then put them
//	 back in any order."
//
// Architects of Will's look-and-reorder, as an activated ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42cffe5d-6e6b-4ac4-b539-36a7b3bc9a65",
		Name:         "Elemental Augury",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: Look at the top three cards of target player's library, then put them back in any order.",
			Cost:    effects.ManaCost("{3}"),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.AugurySetupB831,
		}},
	})
}
