package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tablet of Compleation — Artifact {2}:
//
//	"{T}: Put an oil counter on this artifact.
//	 {T}: Add {C}. Activate only if this artifact has two or more oil
//	 counters on it.
//	 {1}, {T}: Draw a card. Activate only if this artifact has five or more
//	 oil counters on it."
//
// The counter ability uses the stack; the {C} ability is a mana ability
// (CR 605.1a) gated on the live oil count, checked before the tap is paid
// (CR 602.5); the draw is an activated ability with the same kind of gate.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6eca2ac9-16ba-4a11-b6ee-3caf2d7d8254",
		Name:         "Tablet of Compleation",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:      effects.ManaAbilityCost{Tap: true},
			Produced:  "{C}",
			Condition: effects.SourceHasCountersAtLeast(effects.S82OneB12Oil, 2),
			Label:     "{T}: Add {C}. Activate only if this artifact has two or more oil counters on it.",
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Put an oil counter on this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect:  effects.PutACounterOnThis(effects.S82OneB12Oil),
			},
			{
				Label:     "{1}, {T}: Draw a card. Activate only if this artifact has five or more oil counters on it.",
				Purpose:   game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:      effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
				Condition: effects.SourceHasCountersAtLeast(effects.S82OneB12Oil, 5),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return g.DrawNForEffect(item.Controller, 1)
				},
			},
		},
	})
}
