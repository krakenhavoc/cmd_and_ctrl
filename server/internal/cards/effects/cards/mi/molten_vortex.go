package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Molten Vortex — Enchantment {R} (Magic Origins):
//
//	"{R}, Discard a land card: This enchantment deals 2 damage to any
//	 target."
//
// Soul of Windgrace's mana-plus-land-discard cost with Borborygmos
// Enraged's damage. Any number of activations: nothing limits them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da5ee877-143e-4f53-ad20-1bc527a8fff8",
		Name:         "Molten Vortex",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Discard a land card: This enchantment deals 2 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.DiscardCardsMatching(1, "a land card", game.Card.IsLand)),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
