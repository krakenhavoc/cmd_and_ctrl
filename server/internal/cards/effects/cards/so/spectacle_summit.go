package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spectacle Summit — Land (Secrets of Strixhaven, #4456):
//
//	"This land enters tapped.
//	 {T}: Add {U} or {R}.
//	 {2}{U}{R}, {T}: Surveil 1."
//
// Silverquill Campus's shape with a surveil: a nonbasic land with no basic
// types, so both mana colours are declared, and the surveil is an ordinary
// activated ability whose cost taps the land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "33a4e73d-d93a-4b6f-88ff-cd53f20d178c",
		Name:          "Spectacle Summit",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("U", "R")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}{R}, {T}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{U}{R}"), effects.TapCost()),
			Effect:  effects.Do(effects.Surveil{N: 1}),
		}},
	})
}
