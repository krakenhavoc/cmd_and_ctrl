package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of Death's Hold — Enchantment — Aura Curse {3}{B}{B}:
//
//	"Enchant player
//	 Creatures enchanted player controls get -1/-1."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7194cf15-69fc-46f0-987a-2cea8b498f50",
		Name:         "Curse of Death's Hold",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{{
			Label:     "Creatures enchanted player controls get -1/-1.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.OldB444CreaturesEnchantedPlayerControls,
			Apply:     effects.PumpBy(-1, -1),
		}},
	})
}
