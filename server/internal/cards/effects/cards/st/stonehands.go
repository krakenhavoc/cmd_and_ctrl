package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stonehands — Enchantment — Aura {2}{R} (Ice Age, S82 OLD #5636):
//
//	"Enchant creature
//	 Enchanted creature gets +0/+2.
//	 {R}: Enchanted creature gets +1/+0 until end of turn."
//
// The pump is Firebreathing's, on the Aura's own activated ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8beccaff-d30f-4164-a264-cb7119cee4f1",
		Name:         "Stonehands",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(0, 2)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Enchanted creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.S82OldB399PumpEnchanted("Stonehands", 1, 0),
		}},
	})
}
