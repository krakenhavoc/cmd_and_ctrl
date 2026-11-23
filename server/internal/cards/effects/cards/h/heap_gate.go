package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heap Gate — Land — Gate (Commander Legends: Battle for Baldur's Gate, S82
// GAP #5948):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add one mana of any color.
//	 {1}, {T}, Tap an untapped Gate you control: Create a Treasure token."
//
// Two mana abilities (Rootrider Faun's pair) and an activated ability whose
// cost taps another untapped Gate: the Gate is not the {T} symbol (CR 302.6),
// so one that arrived this turn may pay it, and it can never be Heap Gate
// itself, which {T} has already tapped.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "35922a30-6b84-44dd-a2f0-306554a1ae90",
		Name:         "Heap Gate",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "{1}, {T}: Add one mana of any color",
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Tap an untapped Gate you control: Create a Treasure token.",
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.TapAnotherUntapped("an untapped Gate you control", effects.Subtype("Gate"))),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1}),
		}},
	})
}
