package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firebreathing — Enchantment — Aura {R} (Magic 2012, S82 older sets batch
// 451, #5573):
//
//	"Enchant creature
//	 {R}: Enchanted creature gets +1/+0 until end of turn."
//
// Shiv's Embrace's pump ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8603bf74-faab-4910-8e45-0f2e3b318efb",
		Name:         "Firebreathing",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Enchanted creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.S82OldB399PumpEnchanted("Firebreathing", 1, 0),
		}},
	})
}
