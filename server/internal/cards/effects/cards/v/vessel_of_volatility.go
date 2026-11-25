package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vessel of Volatility — {1}{R} Enchantment (Shadows over Innistrad, S82 OLD
// b293):
//
//	"{1}{R}, Sacrifice this enchantment: Add {R}{R}{R}{R}."
//
// Basal Thrull's cash-in without the tap: a mana ability (CR 605.1a) whose
// cost pays {1}{R} and sacrifices the enchantment.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5a4d485-ab9c-48bb-866f-448b1f7707cc",
		Name:         "Vessel of Volatility",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Mana: "{1}{R}", Sacrifice: true},
			Produced: "{R}{R}{R}{R}",
			Label:    "{1}{R}, Sacrifice this enchantment: Add {R}{R}{R}{R}",
		}},
	})
}
