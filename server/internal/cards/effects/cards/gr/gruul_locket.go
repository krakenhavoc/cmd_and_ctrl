package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gruul Locket — Artifact {3} (Ravnica Allegiance, #5521):
//
//	"{T}: Add {R} or {G}.
//	 {R/G}{R/G}{R/G}{R/G}, {T}, Sacrifice this artifact: Draw two cards."
//
// Azorius Locket's two lines in red and green: a mana ability of either
// colour (CR 605.1a) and the four-hybrid, tap-and-sacrifice draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7018ccd0-96e6-458f-b412-197edcb2bafd",
		Name:         "Gruul Locket",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R|G}",
			Label:    "Add {R} or {G}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R/G}{R/G}{R/G}{R/G}, {T}, Sacrifice this artifact: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{R/G}{R/G}{R/G}{R/G}"},
			Effect:  effects.Do(effects.DrawCards{N: 2}),
		}},
	})
}
