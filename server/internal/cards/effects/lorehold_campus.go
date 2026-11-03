package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lorehold Campus — Land (Secrets of Strixhaven Commander, #4834):
//
//	"This land enters tapped.
//	 {T}: Add {R} or {W}.
//	 {4}, {T}: Scry 1."
//
// University Campus's shape with a scry: a nonbasic land with no basic
// types, so both mana colours are declared, and the scry is an ordinary
// activated ability whose cost taps the land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "45773715-3f46-4671-b633-bf087e892e26",
		Name:          "Lorehold Campus",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("R", "W")},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Effect:  Do(Scry{N: 1}),
		}},
	})
}
