package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const staggeringInsightGrant = "staggering-insight/draw"

// Staggering Insight — Enchantment — Aura {W}{U} (Starter Commander Decks):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has lifelink and "Whenever this
//	 creature deals combat damage to a player, draw a card.""
//
// Combat Research's granted draw (a bundle on the enchanted creature, so its
// controller draws), with the pump in layer 7c and lifelink in layer 6.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09071f5f-c9cb-43a9-b41d-faf6caf3beff",
		Name:         "Staggering Insight",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key:       staggeringInsightGrant,
			Triggered: []game.TriggeredAbility{effects.CatgeistDraw("Staggering Insight — draw a card")},
			Text:      "Whenever this creature deals combat damage to a player, draw a card.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("lifelink"),
			effects.GrantAbilitiesToAttached(staggeringInsightGrant),
		},
	})
}
