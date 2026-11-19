package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cartouche of Knowledge — {1}{U} Enchantment — Aura Cartouche:
// "Enchant creature you control. When this Aura enters, draw a card.
// Enchanted creature gets +1/+1 and has flying."
func init() {
	Register(Spec{
		OracleID:     "488696c2-b0c7-49be-a5a1-e833b6fd8751",
		Name:         "Cartouche of Knowledge",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Cartouche of Knowledge — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
