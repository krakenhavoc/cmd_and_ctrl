package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gaea's Embrace — Enchantment — Aura {2}{G}{G} (Urza's Saga, #5624):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3 and has trample.
//	 {G}: Regenerate enchanted creature."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b44e3f2e-3185-40d5-b50c-d3a1e473120c",
		Name:         "Gaea's Embrace",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(3, 3), effects.GrantToAttached("trample")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}: Regenerate enchanted creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{G}"),
			Effect:  effects.RegenerateAttachedHost,
		}},
	})
}
