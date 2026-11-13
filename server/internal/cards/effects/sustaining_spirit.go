package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sustaining Spirit — Creature — Angel Spirit {1}{W}, 0/3 (Alliances, #5770):
//
//	"Cumulative upkeep {1}{W} (At the beginning of your upkeep, put an age
//	 counter on this permanent, then sacrifice it unless you pay its upkeep
//	 cost for each age counter on it.)
//	 Damage that would reduce your life total to less than 1 reduces it to
//	 1 instead."
//
// The shared CumulativeUpkeep trigger (CR 702.24a), and the floor as a
// printed static of the permanent: Spec.PlayerLifeFloorOne, read off the
// battlefield on every query (game/life_floor.go), so it is its
// controller's while it is on the battlefield and nobody's once it leaves.
// The floor caps the life the damage COSTS (CR 120.3a), not the damage:
// lifelink and commander damage see all of it, and "lose life" effects
// and life payments are not floored (the 2008-10-01 rulings).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:           "683c81ae-38cd-4db1-b7e7-a09d34b6431e",
		Name:               "Sustaining Spirit",
		Completeness:       CompletenessFull,
		PlayerLifeFloorOne: true,
		Triggered: []game.TriggeredAbility{
			CumulativeUpkeep("Sustaining Spirit — cumulative upkeep {1}{W}", "{1}{W}"),
		},
	})
}
