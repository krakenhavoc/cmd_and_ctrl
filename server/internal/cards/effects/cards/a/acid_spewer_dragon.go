package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Acid-Spewer Dragon — Creature — Dragon {5}{B}, 3/3 (Dragons of Tarkir,
// #5551):
//
//	"Flying, deathtouch
//	 Megamorph {5}{B}{B} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its megamorph cost and put a +1/+1 counter on it.)
//	 When this creature is turned face up, put a +1/+1 counter on each other Dragon creature you control."
//
// Megamorph is the shared constructor (morph.go): the face-down cast for
// {3}, and the +1/+1 counter placed by the turn-face-up special action
// (CR 702.37b).
// "Each other Dragon creature you control" is read as the trigger
// resolves (Agent Phil Coulson's body), so a Dragon that arrived in
// response gets a counter and this one never does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "f4378682-359f-48b0-adb7-b996b8d972db",
		Name:             "Acid-Spewer Dragon",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying", "deathtouch"},
		AlternativeCosts: []game.AlternativeCost{effects.Megamorph("{5}{B}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisIsTurnedFaceUp("Acid-Spewer Dragon — put a +1/+1 counter on each other Dragon creature you control",
				effects.MshB1CounterOnEachOtherYouControl(effects.And(effects.Creature(), effects.Subtype("Dragon")))),
		},
	})
}
