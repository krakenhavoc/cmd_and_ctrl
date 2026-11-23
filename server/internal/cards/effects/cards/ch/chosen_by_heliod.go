package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chosen by Heliod — {1}{W} Enchantment — Aura: "Enchant creature. When
// this Aura enters, draw a card. Enchanted creature gets +0/+2."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f6475b40-e694-4aac-a25c-bbd48d45127d",
		Name:         "Chosen by Heliod",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Static: []game.StaticAbility{
			effects.PumpAttached(0, 2),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Chosen by Heliod — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
