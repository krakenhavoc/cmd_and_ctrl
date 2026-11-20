package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Font of Fertility — Enchantment {G} (Journey into Nyx, #5558):
//
//	"{1}{G}, Sacrifice this enchantment: Search your library for a
//	 basic land card, put it onto the battlefield tapped, then shuffle."
//
// A sacrifice-as-cost ability, activated at instant speed.
// Evolving Wilds' tapped basic.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7fb6f960-38f6-45fa-bfec-d43fc174daa7",
		Name:         "Font of Fertility",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}, Sacrifice this enchantment: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1},
			Cost:    game.AbilityCost{SacrificeSelf: true, Mana: "{1}{G}"},
			Effect:  fetchBasicTapped,
		}},
	})
}
