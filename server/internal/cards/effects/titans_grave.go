package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Titan's Grave — Land (Secrets of Strixhaven, #4456):
//
//	"This land enters tapped.
//	 {T}: Add {B} or {G}.
//	 {2}{B}{G}, {T}: Surveil 1. (Look at the top card of your library. You
//	 may put it into your graveyard.)"
//
// Paradox Gardens' land in black and green: a nonbasic land with no basic
// types, so both mana colours are declared, and an ordinary activated ability
// whose cost taps the land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "595f0eb5-f521-4174-9c48-b89e85ea907c",
		Name:          "Titan's Grave",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("B", "G")},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}{G}, {T}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{B}{G}"), TapCost()),
			Effect:  Do(Surveil{N: 1}),
		}},
	})
}
