package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Line Cutter — Creature — Human Rogue Guest {3}{B}:
//
//	"When this creature enters, roll to visit your Attractions."
//
// The roll is CR 701.52a's, outside the precombat main phase: a d6 (an
// ordinary die roll, so "whenever you roll" sees it), then every
// Attraction its controller controls with that number lit up is visited.
// With no Attractions it still rolls and visits nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a43d9793-b2ee-427c-95ca-285611aa1a4d",
		Name:         "Line Cutter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Line Cutter — roll to visit your Attractions", Do(RollToVisitYourAttractions{})),
		},
	})
}
