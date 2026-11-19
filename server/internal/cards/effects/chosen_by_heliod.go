package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chosen by Heliod — {1}{W} Enchantment — Aura: "Enchant creature. When
// this Aura enters, draw a card. Enchanted creature gets +0/+2."
func init() {
	Register(Spec{
		OracleID:     "f6475b40-e694-4aac-a25c-bbd48d45127d",
		Name:         "Chosen by Heliod",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Static: []game.StaticAbility{
			PumpAttached(0, 2),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Chosen by Heliod — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
