package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Realm of Koh — Land (Avatar: The Last Airbender, #4459):
//
//	"This land enters tapped unless you control a basic land.
//	 {T}: Add {B}.
//	 {3}{B}, {T}: Create a 1/1 colorless Spirit creature token with
//	 "This token can't block or be blocked by non-Spirit creatures.""
//
// The black member of the set's "abandoned temple" cycle (Abandoned Air
// Temple): the tapped entry is the EntersTappedUnless self-replacement on
// b11ControlsBasicLand, read as the land enters, so a basic land entering
// at the same time is not counted (the ruling). The Spirit is Avatar
// Kuruk's printed token, block rule included (kurukSpiritToken).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bb9ce416-eef1-49e8-89a0-2b6837505070",
		Name:         "Realm of Koh",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{EntersTappedUnless(b11ControlsBasicLand)},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}{B}, {T}: Create a 1/1 colorless Spirit creature token with \"This token can't block or be blocked by non-Spirit creatures.\"",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
			Cost:    Plus(ManaCost("{3}{B}"), TapCost()),
			Effect:  Do(CreateToken{Template: kurukSpiritToken(), N: 1}),
		}},
	})
}
