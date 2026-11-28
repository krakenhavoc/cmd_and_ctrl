package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mourning — Enchantment — Aura {1}{B} (Invasion, S82 older sets batch 661, #5616):
//
//	"Enchant creature
//	 Enchanted creature gets -2/-0.
//	 {B}: Return this Aura to its owner's hand."
//
// Conviction's return ability with a static on the enchanted creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "20602d11-c0ae-47ce-a7fd-998b12480e18",
		Name:         "Mourning",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.PumpAttached(-2, 0)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: Return this Aura to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{B}"),
			Effect:  effects.Do(effects.Instruction(effects.StdB9ReturnThisToHand)),
		}},
	})
}
