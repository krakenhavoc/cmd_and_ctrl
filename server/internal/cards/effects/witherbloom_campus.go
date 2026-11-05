package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Witherbloom Campus — Land (Secrets of Strixhaven Commander, #4834):
//
//	"This land enters tapped.
//	 {T}: Add {B} or {G}.
//	 {4}, {T}: Scry 1."
//
// Lorehold Campus's shape in black and green: a nonbasic land with no basic
// types, so both mana colours are declared, and the scry is an ordinary
// activated ability whose cost taps the land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "6cd58a88-6434-4c55-bf93-a739b5ed9bc1",
		Name:          "Witherbloom Campus",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("B", "G")},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Effect:  Do(Scry{N: 1}),
		}},
	})
}
