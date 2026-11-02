package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ominous Asylum — Land (Marvel's Spider-Man):
//
//	"This land enters tapped.
//	 {T}: Add {B} or {R}.
//	 {4}, {T}: Surveil 1."
//
// The tapped entry is the real CR 614 self-replacement (Abraded Bluffs);
// the surveil is the shared Surveil primitive behind a mana-and-tap
// activation.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "676141c3-a433-4aba-86fb-729628f96dfa",
		Name:         "Ominous Asylum",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B|R}",
			Label:    "Add {B} or {R}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Effect:  Do(Surveil{N: 1}),
		}},
	})
}
