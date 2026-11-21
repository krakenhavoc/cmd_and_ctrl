package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sheltering Boughs — {2}{G} Enchantment — Aura (Innistrad: Crimson Vow,
// S82):
//
//	"Enchant creature
//	 When this Aura enters, draw a card.
//	 Enchanted creature gets +1/+3."
//
// Angelic Gift's enters-draw with PumpAttached for the layer 7c change.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6bde4eb-15f0-4e6e-9d25-459a2f75ccf2",
		Name:         "Sheltering Boughs",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Static:       []game.StaticAbility{effects.PumpAttached(1, 3)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sheltering Boughs — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
