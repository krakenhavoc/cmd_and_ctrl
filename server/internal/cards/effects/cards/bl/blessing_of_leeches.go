package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blessing of Leeches — Enchantment — Aura {2}{B} (Betrayers of Kamigawa,
// #5602):
//
//	"Flash
//	 Enchant creature
//	 At the beginning of your upkeep, you lose 1 life.
//	 {0}: Regenerate enchanted creature."
//
// Flash rides PrintedKeywords. The upkeep trigger is the Aura
// controller's ("you"), whoever controls the creature. The regeneration
// is Dark Privilege's regenerateAttachedHost with a {0} cost, so it can
// be activated any number of times; each makes one shield (CR 701.19a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "923b725a-e189-4b61-b01a-c5056f1328ba",
		Name:            "Blessing of Leeches",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Blessing of Leeches — you lose 1 life", effects.Do(effects.GainLife{Amount: -1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{0}: Regenerate enchanted creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{0}"),
			Effect:  effects.RegenerateAttachedHost,
		}},
	})
}
