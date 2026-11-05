package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prismari Campus — Land (Secrets of Strixhaven Commander, #4834):
//
//	"This land enters tapped.
//	 {T}: Add {U} or {R}.
//	 {4}, {T}: Scry 1."
//
// University Campus's shape with a scry in place of the surveil: a nonbasic
// land with no basic types, so both mana colours are declared, and an
// ordinary activated ability whose cost taps the land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "3a3a1b35-ae4d-49d5-ae09-5a1693ad53ce",
		Name:          "Prismari Campus",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("U", "R")},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Effect:  Do(Scry{N: 1}),
		}},
	})
}
