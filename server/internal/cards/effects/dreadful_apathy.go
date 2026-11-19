package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     dreadfulApathyOracle,
		Name:         "Dreadful Apathy",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{RestrictAttached(game.CantAttackOrBlock)},
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}: Exile enchanted creature.",
			Cost:    ManaCost("{2}{W}"),
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect:  coopedUpExileEnchanted,
		}},
	})
}
