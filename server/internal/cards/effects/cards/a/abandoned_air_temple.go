package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abandoned Air Temple — Land (EDHREC rank 1231):
//
//	"This land enters tapped unless you control a basic land.
//	 {T}: Add {W}.
//	 {3}{W}, {T}: Put a +1/+1 counter on each creature you control."
//
// The Avatar: The Last Airbender "abandoned temple" cycle: a Castle
// Ardenvale whose gate is any basic land rather than a Plains, and
// whose activation is a Gavony Township for four. The tapped entry is
// the EntersTappedUnless self-replacement on the b11ControlsBasicLand
// condition; the activation is a CR 602 ability with a mana-and-tap
// cost whose effect snapshots the creatures first, so a token made
// by a counter-placement trigger mid-loop does not receive one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9575d7ce-f26d-4b90-87a3-6329e9799572",
		Name:         "Abandoned Air Temple",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.EntersTappedUnless(effects.B11ControlsBasicLand)},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}, {T}: Put a +1/+1 counter on each creature you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{3}{W}"), effects.TapCost()),
			Effect:  effects.B11PutCounterOnEachCreatureYouControl,
		}},
	})
}
