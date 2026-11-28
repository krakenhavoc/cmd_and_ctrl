package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armor of Faith — Enchantment — Aura {W} (Ice Age, S82 OLD b828, #5636):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1.
//	 {W}: Enchanted creature gets +0/+1 until end of turn."
//
// PumpAttached for the static and Firebreathing's pump ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c482b20-b423-48ef-b162-53960d43b4f7",
		Name:         "Armor of Faith",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}: Enchanted creature gets +0/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{W}"),
			Effect:  effects.S82OldB399PumpEnchanted("Armor of Faith", 0, 1),
		}},
	})
}
