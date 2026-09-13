package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "9575d7ce-f26d-4b90-87a3-6329e9799572",
		Name:         "Abandoned Air Temple",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{EntersTappedUnless(b11ControlsBasicLand)},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []ActivatedAbility{{
			Label:  "{3}{W}, {T}: Put a +1/+1 counter on each creature you control.",
			Cost:   Plus(ManaCost("{3}{W}"), TapCost()),
			Effect: b11PutCounterOnEachCreatureYouControl,
		}},
	})
}
