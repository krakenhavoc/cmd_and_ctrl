package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Simic Locket — Artifact {3} (Ravnica Allegiance, #5521):
//
//	"{T}: Add {G} or {U}.
//	 {G/U}{G/U}{G/U}{G/U}, {T}, Sacrifice this artifact: Draw two cards."
//
// Azorius Locket's two lines in green and blue: a mana ability of either
// colour (CR 605.1a) and the four-hybrid, tap-and-sacrifice draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c567fed8-df00-4579-901e-8754079105bf",
		Name:         "Simic Locket",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G|U}",
			Label:    "Add {G} or {U}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{G/U}{G/U}{G/U}{G/U}, {T}, Sacrifice this artifact: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{G/U}{G/U}{G/U}{G/U}"},
			Effect:  Do(DrawCards{N: 2}),
		}},
	})
}
