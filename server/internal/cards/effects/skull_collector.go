package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skull Collector — Creature — Ogre Warrior {1}{B}{B}, 3/3 (Saviors of Kamigawa, #5601):
//
//	"At the beginning of your upkeep, return a black creature you control to its owner's hand.
//	 {1}{B}: Regenerate this creature."
//
// The upkeep return is the shared sokB1ReturnACreatureOfColour: not
// targeted and not optional, the black creature picked as it resolves,
// this creature itself included. The second row is the shared regenerateThisRow (CR 701.19a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "42ec62b1-c899-4ebd-b786-28edf76ce8e0",
		Name:         "Skull Collector",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{sokB1ReturnACreatureOfColour("Skull Collector", "B", "black")},
		Activated:    []ActivatedAbility{regenerateThisRow("{1}{B}")},
	})
}
