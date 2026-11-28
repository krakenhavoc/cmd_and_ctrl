package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jolrael's Favor — Enchantment — Aura {1}{G} (Prophecy, S82 OLD b669, #5617):
//
//	"Flash
//	 Enchant creature
//	 {1}{G}: Regenerate enchanted creature."
//
// Blessing of Leeches' RegenerateAttachedHost with a {1}{G} cost: the ability is
// the Aura's own, so its controller pays and the shield goes on whatever it is
// attached to.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bf71cca7-be71-4429-9cb6-1eb34ff847eb",
		Name:            "Jolrael's Favor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}: Regenerate enchanted creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}{G}"),
			Effect:  effects.RegenerateAttachedHost,
		}},
	})
}
