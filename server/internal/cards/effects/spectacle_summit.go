package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "33a4e73d-d93a-4b6f-88ff-cd53f20d178c",
		Name:          "Spectacle Summit",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("U", "R")},
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}{R}, {T}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{U}{R}"), TapCost()),
			Effect:  Do(Surveil{N: 1}),
		}},
	})
}
