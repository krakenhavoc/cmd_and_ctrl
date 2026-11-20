package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Font of Fortunes — Enchantment {1}{U} (Journey into Nyx, #5558):
//
//	"{1}{U}, Sacrifice this enchantment: Draw two cards."
//
// A sacrifice-as-cost ability, activated at instant speed.
// Azorius Locket's draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4aeaf1cd-7346-4492-847d-1c4dad14fafd",
		Name:         "Font of Fortunes",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}, Sacrifice this enchantment: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{SacrificeSelf: true, Mana: "{1}{U}"},
			Effect:  Do(DrawCards{N: 2}),
		}},
	})
}
