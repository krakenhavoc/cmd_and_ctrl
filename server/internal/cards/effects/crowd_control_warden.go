package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crowd-Control Warden — Creature — Centaur Soldier {3}{G}{W}, 4/4
// (Murders at Karlov Manor, #4469):
//
//	"As this creature enters or is turned face up, put X +1/+1 counters
//	 on it, where X is the number of other creatures you control.
//	 Disguise {3}{G/W}{G/W} (You may cast this card face down for {3} as
//	 a 2/2 creature with ward {2}. Turn it face up any time for its
//	 disguise cost.)"
//
// One clause, two replacement effects: "as this enters" is a CR 614.1c
// self-entry replacement counted as it enters, and "as this is turned
// face up" is a CR 614.1e replacement applied while it turns face up
// (CR 708.11, #5237). Both count the board at that moment. Entering
// face down it has no abilities (CR 708.2a), so it enters with no
// counters, and turning it face up later is not entering (CR 708.8), so
// it gets the face-up count once and not both.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "439e59ab-0350-4729-9352-60e697eaa78d",
		Name:         "Crowd-Control Warden",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted("+1/+1", crowdControlWardenX,
				"Crowd-Control Warden: enters with a +1/+1 counter per other creature you control"),
		},
		AlternativeCosts: []game.AlternativeCost{
			AsTurnedFaceUp(Disguise("{3}{G/W}{G/W}"), game.AsTurnedFaceUpCounters{Count: crowdControlWardenX}),
		},
	})
}

// crowdControlWardenX is "the number of other creatures you control",
// for the Warden `src`: as it enters (it is not on the battlefield yet)
// and as it is turned face up (it is, and is left out).
func crowdControlWardenX(g *game.Game, src *game.Card) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != src.InstanceID && c.Controller == src.Controller && c.IsCreature() {
			n++
		}
	}
	return n
}
