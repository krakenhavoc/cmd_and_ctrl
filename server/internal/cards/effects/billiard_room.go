package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Billiard Room — Land (Ravnica: Clue Edition):
//
//	"This land enters tapped.
//	 {T}: Add {B} or {R}.
//	 {4}, {T}: Investigate."
//
// Ominous Asylum's shape with an investigate in place of the surveil: the
// real CR 614 self-replacement for the tapped entry, a mana ability, and a
// mana-and-tap activation that creates a Clue for its controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a4fc174e-7fa6-41a8-ae03-255f226840f9",
		Name:         "Billiard Room",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B|R}",
			Label:    "Add {B} or {R}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Investigate.",
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Effect:  Do(Investigate{N: 1}),
		}},
	})
}
