package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Belltoll Dragon — Creature — Dragon {5}{U}, 3/3 (Dragons of Tarkir,
// #5551):
//
//	"Flying, hexproof
//	 Megamorph {5}{U}{U} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its megamorph cost and put a +1/+1 counter on it.)
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
	Register(Spec{
		OracleID:         "d0e529e3-0f77-4a55-ba35-46591f48d960",
		Name:             "Belltoll Dragon",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying", "hexproof"},
		AlternativeCosts: []game.AlternativeCost{Megamorph("{5}{U}{U}")},
		Triggered: []game.TriggeredAbility{
			WhenThisIsTurnedFaceUp("Belltoll Dragon — put a +1/+1 counter on each other Dragon creature you control",
				mshB1CounterOnEachOtherYouControl(And(Creature(), Subtype("Dragon")))),
		},
	})
}
