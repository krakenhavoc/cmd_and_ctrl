package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Titan's Grave — Land (Secrets of Strixhaven, #4456):
//
//	"This land enters tapped.
//	 {T}: Add {B} or {G}.
//	 {2}{B}{G}, {T}: Surveil 1. (Look at the top card of your library. You
//	 may put it into your graveyard.)"
//
// Paradox Gardens' land in black and green: a nonbasic land with no basic
// types, so both mana colours are declared, and an ordinary activated ability
// whose cost taps the land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "595f0eb5-f521-4174-9c48-b89e85ea907c",
		Name:          "Titan's Grave",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("B", "G")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}{G}, {T}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}{G}"), effects.TapCost()),
			Effect:  effects.Do(effects.Surveil{N: 1}),
		}},
	})
}
