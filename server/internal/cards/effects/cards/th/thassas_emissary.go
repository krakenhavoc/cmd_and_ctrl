package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thassa's Emissary — Enchantment Creature — Crab {3}{U}, 3/3 (Theros):
//
//	"Bestow {5}{U} (If you cast this card for its bestow cost, it's an Aura
//	 spell with enchant creature. It becomes a creature again if it's not
//	 attached.)
//	 Whenever this creature or enchanted creature deals combat damage to a
//	 player, draw a card.
//	 Enchanted creature gets +3/+3."
//
// One printed trigger: combat damage to a player dealt by the source or by its
// host (OldB395ThisOrEnchantedDealtCombatDamageToAPlayer). Each damage event
// draws once, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "fcb97d06-e0b8-4a13-ab7c-1304ccb7ace2",
		Name:             "Thassa's Emissary",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Draws: 1},
		AlternativeCosts: []game.AlternativeCost{effects.Bestow("{5}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.OldB395ThisOrEnchantedDealtCombatDamageToAPlayer,
				"Thassa's Emissary — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Static: []game.StaticAbility{effects.PumpAttached(3, 3)},
	})
}
