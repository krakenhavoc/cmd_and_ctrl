package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cartouche of Knowledge — {1}{U} Enchantment — Aura Cartouche:
// "Enchant creature you control. When this Aura enters, draw a card.
// Enchanted creature gets +1/+1 and has flying."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "488696c2-b0c7-49be-a5a1-e833b6fd8751",
		Name:         "Cartouche of Knowledge",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cartouche of Knowledge — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
