package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fields of Strife — Land (Secrets of Strixhaven, #4456):
//
//	"This land enters tapped.
//	 {T}: Add {R} or {W}.
//	 {2}{R}{W}, {T}: Surveil 1. (Look at the top card of your library. You
//	 may put it into your graveyard.)"
//
// Ominous Asylum's land (a tapped dual with a surveil sink), in red and white
// with a four-mana activation that needs {R}{W} of it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a32e08fa-bea4-4ba9-a126-9bf0a91f67e2",
		Name:         "Fields of Strife",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R|W}",
			Label:    "Add {R} or {W}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}{W}, {T}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{R}{W}"), TapCost()),
			Effect:  Do(Surveil{N: 1}),
		}},
	})
}
