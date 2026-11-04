package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quandrix Campus — Land (Secrets of Strixhaven Commander, #4834):
//
//	"This land enters tapped.
//	 {T}: Add {G} or {U}.
//	 {4}, {T}: Scry 1."
//
// Prismari Campus's shape in other colours: a nonbasic land with no basic
// types, so both mana colours are declared, and an ordinary activated ability
// whose cost taps the land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "172f86b6-9580-4eb2-b7dc-2a44277d978b",
		Name:          "Quandrix Campus",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("G", "U")},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Effect:  Do(Scry{N: 1}),
		}},
	})
}
