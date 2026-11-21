package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "c567fed8-df00-4579-901e-8754079105bf",
		Name:         "Simic Locket",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G|U}",
			Label:    "Add {G} or {U}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G/U}{G/U}{G/U}{G/U}, {T}, Sacrifice this artifact: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{G/U}{G/U}{G/U}{G/U}"},
			Effect:  effects.Do(effects.DrawCards{N: 2}),
		}},
	})
}
