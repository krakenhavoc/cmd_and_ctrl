package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cowed by Wisdom — Enchantment — Aura {W} (Saviors of Kamigawa, #5601):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block unless its controller pays {1}
//	 for each card in your hand."
//
// Oppressive Rays' attack and block taxes (CR 508.1h, 509.1d) at a price of
// {1} for each card in the Aura's controller's hand, counted as the attack or
// block is declared. The creature's controller pays, whoever owns the Aura; an
// empty hand charges nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10495c22-2456-4b93-afa9-0c3c53075efe",
		Name:         "Cowed by Wisdom",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		AttackTaxes: []game.AttackTax{{
			Label:      "Enchanted creature can't attack unless its controller pays {1} for each card in your hand.",
			OnAttacker: true,
			AppliesTo:  func(q game.AttackTaxQuery) bool { return q.Source.IsAttachedTo(q.Attacker.InstanceID) },
			ManaCost: func(q game.AttackTaxQuery) string {
				return effects.GenericMana(effects.CardsInHand(q.Game, q.Source.Controller))
			},
		}},
		BlockTaxes: []game.BlockTax{{
			Label:     "Enchanted creature can't block unless its controller pays {1} for each card in your hand.",
			AppliesTo: func(q game.BlockTaxQuery) bool { return q.Source.IsAttachedTo(q.Blocker.InstanceID) },
			ManaCost: func(q game.BlockTaxQuery) string {
				return effects.GenericMana(effects.CardsInHand(q.Game, q.Source.Controller))
			},
		}},
	})
}
