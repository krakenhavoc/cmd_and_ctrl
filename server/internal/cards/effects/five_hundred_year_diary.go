package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Five Hundred Year Diary — Legendary Artifact — Book Clue {3}{U} (Doctor Who,
// #5948):
//
//	"Five Hundred Year Diary enters tapped.
//	 {T}: Add {U} for each Clue you control.
//	 {2}, Sacrifice Five Hundred Year Diary: Draw a card."
//
// The Diary is itself a Clue, so it counts itself. The mana ability counts
// Clues as it resolves (Gaea's Cradle's scaled shape); with none it adds
// nothing, which cannot happen while the Diary is on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9bdd6fea-d436-412b-a7b7-bbdba73dc986",
		Name:         "Five Hundred Year Diary",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:           ManaAbilityCost{Tap: true},
			ProducedAmount: ProducedPerPermanent("U", func(c game.Card) bool { return c.HasSubtype("Clue") }),
			Label:          "Add {U} for each Clue you control",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice Five Hundred Year Diary: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Mana: "{2}", SacrificeSelf: true},
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
