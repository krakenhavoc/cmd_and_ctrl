package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magus of the Library — Creature — Human Wizard {G}{G}, 1/1:
//
//	"{T}: Add {C}.
//	 {T}: Draw a card. Activate only if you have exactly seven cards in
//	 hand."
//
// A plain {C} mana ability, and Resonating Lute's draw with its "activate
// only if" condition made exact: the hand must hold seven cards, no more and
// no fewer, when the ability is activated (the {T} cost is paid after the
// check, so the hand is unchanged by it).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69266cb7-50eb-43d2-96d6-fc3c85be49f7",
		Name:         "Magus of the Library",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:     "{T}: Draw a card. Activate only if you have exactly seven cards in hand.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      TapCost(),
			Condition: s82Bns04HandSizeExactly(7),
			Effect:    Do(DrawCards{N: 1}),
		}},
	})
}
