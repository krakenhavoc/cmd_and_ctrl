package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Douse — Enchantment {2}{U} (Urza's Saga):
//
//	"{1}{U}: Counter target red spell."
//
// Tidal Control's counter, for a red spell and only for its controller.
// The colour is judged at activation and again as the ability resolves
// (CR 608.2b), so a spell that stopped being red is not countered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3bdc96d9-7272-4e38-b81a-ec928dd49eaa",
		Name:         "Douse",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: Counter target red spell.",
			Cost:    effects.ManaCost("{1}{U}"),
			Targets: effects.TargetSpell("target red spell", effects.OfColor("R")),
			Effect:  effects.CounterTheChosenSpell,
		}},
	})
}
