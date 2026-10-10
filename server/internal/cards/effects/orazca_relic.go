package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Orazca Relic — Artifact {3}:
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 {T}: Add {C}.
//	 {T}, Sacrifice this artifact: You gain 3 life and draw a card.
//	 Activate only if you have the city's blessing."
//
// The city's blessing is the player designation ascend gives and keeps
// (CR 702.131, #2696). The Relic earns it itself the moment its
// controller has ten permanents (it counts as one of them), so the
// sacrifice ability is available from then on, including after the
// Relic or the rest of the board is gone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "48b84b58-1a06-4bb4-be1f-ad3ca69e66dc",
		Name:            "Orazca Relic",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:     "{T}, Sacrifice this artifact: You gain 3 life and draw a card. Activate only if you have the city's blessing.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      Plus(TapCost(), SacrificeThis()),
			Condition: YouHaveTheCitysBlessingCondition(),
			Effect:    Do(GainLife{Amount: 3}, DrawCards{N: 1}),
		}},
	})
}
