package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bubble Smuggler — Creature — Octopus Fish {1}{U}, 2/1 (Murders at
// Karlov Manor, #4469):
//
//	"Disguise {5}{U} (You may cast this card face down for {3} as a 2/2
//	 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)
//	 As this creature is turned face up, put four +1/+1 counters on it."
//
// The disguise keyword (CR 702.168a), and the four counters as a
// CR 614.1e "as this is turned face up" replacement (#5237): they go on
// while it turns face up (CR 708.11), however it comes up, so it arrives
// a 6/5. Cast face up it is a plain 2/1, because the clause only applies
// to turning face up.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ca22fa33-5460-4d31-a9c8-4115d35b994c",
		Name:         "Bubble Smuggler",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			AsTurnedFaceUp(Disguise("{5}{U}"), game.AsTurnedFaceUpCounters{N: 4}),
		},
	})
}
