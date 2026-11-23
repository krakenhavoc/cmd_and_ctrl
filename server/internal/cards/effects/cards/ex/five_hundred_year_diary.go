package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "9bdd6fea-d436-412b-a7b7-bbdba73dc986",
		Name:         "Five Hundred Year Diary",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:           effects.ManaAbilityCost{Tap: true},
			ProducedAmount: effects.ProducedPerPermanent("U", func(c game.Card) bool { return c.HasSubtype("Clue") }),
			Label:          "Add {U} for each Clue you control",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice Five Hundred Year Diary: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Mana: "{2}", SacrificeSelf: true},
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
