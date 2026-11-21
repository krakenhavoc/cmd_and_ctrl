package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ballroom — Land (Ravnica: Clue Edition):
//
//	"This land enters tapped.
//	 {T}: Add {W} or {B}.
//	 {4}, {T}: Investigate."
//
// Ominous Asylum's shape with an investigate in place of the surveil: the
// real CR 614 self-replacement for the tapped entry, a mana ability, and a
// mana-and-tap activation that creates a Clue for its controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb91e842-9f06-4863-a328-2cabe1bcfe27",
		Name:         "Ballroom",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|B}",
			Label:    "Add {W} or {B}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Investigate.",
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Effect:  effects.Do(effects.Investigate{N: 1}),
		}},
	})
}
