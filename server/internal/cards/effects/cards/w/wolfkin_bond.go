package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wolfkin Bond — Enchantment — Aura {4}{G} (Eldritch Moon, #5541):
//
//	"Enchant creature
//	 When this Aura enters, create a 2/2 green Wolf creature token.
//	 Enchanted creature gets +2/+2."
//
// The Wolf is the token table's row; the pump is +2 Mace's on the host.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49e22e88-d4d9-4781-81e3-bf77b76b132b",
		Name:         "Wolfkin Bond",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Wolfkin Bond — create a 2/2 green Wolf creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: 1})),
		},
	})
}
