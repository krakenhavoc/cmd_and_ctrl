package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rakdos Locket — Artifact {3} (Ravnica Allegiance, #5521):
//
//	"{T}: Add {B} or {R}.
//	 {B/R}{B/R}{B/R}{B/R}, {T}, Sacrifice this artifact: Draw two cards."
//
// Azorius Locket's two lines in black and red: a mana ability of either
// colour (CR 605.1a) and the four-hybrid, tap-and-sacrifice draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab784412-0722-4501-9aa0-fd68d307f6fd",
		Name:         "Rakdos Locket",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B|R}",
			Label:    "Add {B} or {R}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B/R}{B/R}{B/R}{B/R}, {T}, Sacrifice this artifact: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{B/R}{B/R}{B/R}{B/R}"},
			Effect:  effects.Do(effects.DrawCards{N: 2}),
		}},
	})
}
