package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Manacles of Decay — Enchantment — Aura {1}{W} (Apocalypse, S82 OLD, #5614):
//
//	"Enchant creature
//	 Enchanted creature can't attack.
//	 {B}: Enchanted creature gets -1/-1 until end of turn.
//	 {R}: Enchanted creature can't block this turn."
//
// Weight of Conscience's attack restriction, and Blessing's Aura ability read
// off the creature the Aura is attached to as it resolves (the Aura's
// controller pays, whoever controls the creature); nothing happens if it is no
// longer attached (CR 608.2b). Each activation is its own effect.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69095c25-6ae6-42f5-9c3d-1d8f86f03fae",
		Name:         "Manacles of Decay",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.RestrictAttached(game.CantAttack)},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{B}: Enchanted creature gets -1/-1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.ManaCost("{B}"),
				Effect:  effects.S82OldB399PumpEnchanted("Manacles of Decay", -1, -1),
			},
			{
				Label:   "{R}: Enchanted creature can't block this turn.",
				Purpose: game.Purpose{Answers: game.AnswerRestrict},
				Cost:    effects.ManaCost("{R}"),
				Effect:  effects.OldB631EnchantedCantBlock("Manacles of Decay — can't block this turn"),
			},
		},
	})
}
