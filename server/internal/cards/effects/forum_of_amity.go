package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Forum of Amity — Land (Secrets of Strixhaven, #4456):
//
//	"This land enters tapped.
//	 {T}: Add {W} or {B}.
//	 {2}{W}{B}, {T}: Surveil 1."
//
// Silverquill Campus's shape with a surveil: a nonbasic land with no basic
// types, so both mana colours are declared, and the surveil is an ordinary
// activated ability whose cost taps the land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "349ea6c7-6b3e-417f-b082-b712e2b1635b",
		Name:          "Forum of Amity",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("W", "B")},
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}{B}, {T}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{W}{B}"), TapCost()),
			Effect:  Do(Surveil{N: 1}),
		}},
	})
}
