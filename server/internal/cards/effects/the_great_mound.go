package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Great Mound — Land (Marvel Super Heroes Commander, #4455):
//
//	"{T}: Add {C}.
//	 {3}, {T}: Create a tapped Vibranium token. (It's an artifact with
//	 indestructible and "{T}: Add {C}. This mana can't be spent to cast a
//	 nonartifact spell.")
//	 {6}, {T}: Draw a card."
//
// The mana is an ordinary tap ability. The second is Dora Milaje Elite's
// printed Vibranium token, created tapped; the third a plain draw. Each
// paid ability taps the land as a cost, so the {C} ability is not
// available the same turn the others are used.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6b78e417-44f8-4ab6-9d3c-e71704fc648e",
		Name:         "The Great Mound",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "{3}, {T}: Create a tapped Vibranium token.",
				Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{3}"), TapCost()),
				Effect:  Do(CreateTokenAdvanced{Spec: Token(VibraniumToken()).EntersTapped(), N: 1}),
			},
			{
				Label:   "{6}, {T}: Draw a card.",
				Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{6}"), TapCost()),
				Effect:  Do(DrawCards{N: 1}),
			},
		},
	})
}
