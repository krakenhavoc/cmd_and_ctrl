package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// University Campus — Land (Marvel's Spider-Man, #4460):
//
//	"This land enters tapped.
//	 {T}: Add {W} or {U}.
//	 {4}, {T}: Surveil 1."
//
// A nonbasic land with no basic types, so both mana colours are declared. The
// surveil is an ordinary activated ability whose cost taps the land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "98e547de-b963-4ee4-9a08-67bae010734b",
		Name:          "University Campus",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("W", "U")},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Effect:  Do(Surveil{N: 1}),
		}},
	})
}
