package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serra's Sanctum — Legendary Land (Urza's Saga, S82 OLD batch 731, #5624):
//
//	"{T}: Add {W} for each enchantment you control."
//
// A mana ability (CR 605.1a) whose amount is counted as it resolves, through
// the layers: every enchantment its controller controls, itself excepted
// because it is a land. With none it adds nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "34187c71-6033-4058-aadc-2bc266f762be",
		Name:         "Serra's Sanctum",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:           effects.ManaAbilityCost{Tap: true},
			ProducedAmount: effects.ProducedPerPermanent("W", effects.MatchEnchantment),
			Label:          "Add {W} for each enchantment you control",
		}},
	})
}
