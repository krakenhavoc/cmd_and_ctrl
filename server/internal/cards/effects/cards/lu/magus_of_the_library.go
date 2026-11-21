package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "69266cb7-50eb-43d2-96d6-fc3c85be49f7",
		Name:         "Magus of the Library",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: Draw a card. Activate only if you have exactly seven cards in hand.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.TapCost(),
			Condition: effects.S82Bns04HandSizeExactly(7),
			Effect:    effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
