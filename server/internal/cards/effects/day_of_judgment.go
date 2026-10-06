package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Day of Judgment — "Destroy all creatures." Wrath of God without
// the "can't be regenerated" clause, which since #667 is a real
// difference: a creature with a regeneration shield survives this and
// not a Wrath. Shares the wrathDestroyAllCreatures helper defined in
// wrath_of_god.go with the other riderless printings.
func init() {
	Register(Spec{
		OracleID:     "d057289d-5e28-43d5-8ff3-4a1bc723477d",
		Name:         "Day of Judgment",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Completeness: CompletenessFull,
		OnResolve:    wrathDestroyAllCreatures,
	})
}
