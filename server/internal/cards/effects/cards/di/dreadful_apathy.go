package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dreadful Apathy — Enchantment — Aura {2}{W} (Foundations Jumpstart):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block.
//	 {2}{W}: Exile enchanted creature."
//
// Cooped Up's text, so it shares its exile body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     effects.DreadfulApathyOracle,
		Name:         "Dreadful Apathy",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.RestrictAttached(game.CantAttackOrBlock)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}: Exile enchanted creature.",
			Cost:    effects.ManaCost("{2}{W}"),
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect:  effects.CoopedUpExileEnchanted,
		}},
	})
}
