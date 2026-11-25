package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sixth Sense — Enchantment — Aura {G} (Amonkhet, S82 OLD batch 244, #5535):
//
//	"Enchant creature
//	 Enchanted creature has "Whenever this creature deals combat damage to a
//	 player, you may draw a card.""
//
// An ability grant of Combat Research's trigger with a "you may": the
// creature's controller is asked as the ability resolves, and a "No" draws
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25d35e3e-ece8-4c94-9dbc-d076a52fca0d",
		Name:         "Sixth Sense",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: sixthSenseGrant,
			Triggered: []game.TriggeredAbility{
				effects.Optional(effects.CatgeistDraw("Sixth Sense — you may draw a card"),
					"Sixth Sense — draw a card?"),
			},
			Text: "Whenever this creature deals combat damage to a player, you may draw a card.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(sixthSenseGrant)},
	})
}

const sixthSenseGrant = "sixth-sense/may-draw"
