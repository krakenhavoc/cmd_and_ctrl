package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mantle of the Wolf — Enchantment — Aura {3}{G}:
//
//	"Enchant creature
//	 Enchanted creature gets +4/+4.
//	 When this Aura is put into a graveyard from the battlefield, create two
//	 2/2 green Wolf creature tokens."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd69b2c7-2be7-4915-9d17-4fd400ace6b8",
		Name:         "Mantle of the Wolf",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Tokens: 2},
		Static:       []game.StaticAbility{effects.PumpAttached(4, 4)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Mantle of the Wolf — create two 2/2 green Wolf creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: 2})),
		},
	})
}
