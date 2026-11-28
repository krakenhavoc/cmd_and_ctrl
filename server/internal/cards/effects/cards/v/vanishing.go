package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vanishing — Enchantment — Aura {U} (Visions, S82 older sets batch 790,
// #5631):
//
//	"Enchant creature
//	 {U}{U}: Enchanted creature phases out. (While it's phased out, it's
//	 treated as though it doesn't exist. It phases in before its controller
//	 untaps during their next untap step.)"
//
// Robe of Stars' ability on an Aura: the Aura's controller pays, whoever
// controls the creature, and the Aura goes with its host (CR 702.26g). An Aura
// that has fallen off by resolution phases nothing out.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd2791e8-0b61-49b8-8b72-ff2737df1049",
		Name:         "Vanishing",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}{U}: Enchanted creature phases out.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ManaCost("{U}{U}"),
			Effect:  effects.PhaseOutAttachedHost,
		}},
	})
}
