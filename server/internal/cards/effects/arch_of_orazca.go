package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arch of Orazca — Land (EDHREC rank 2395):
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 {T}: Add {C}.
//	 {5}, {T}: Draw a card. Activate only if you have the city's
//	 blessing."
//
// A colorless land that draws in a wide deck. "Activate only if you
// have the city's blessing" is the draw's activation condition (CR
// 602.1b, #743).
//
// The city's blessing is the player designation ascend gives and keeps
// (CR 702.131, #2696): the land earns it itself, as soon as its
// controller has ten permanents, and the draw stays open after the
// board shrinks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3bb518ff-399b-4ce7-b9ad-a1d563dd7792",
		Name:            "Arch of Orazca",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:     "{5}, {T}: Draw a card. Activate only if you have the city's blessing.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      Plus(ManaCost("{5}"), TapCost()),
			Condition: YouHaveTheCitysBlessingCondition(),
			Effect:    b36DrawOne,
		}},
	})
}
