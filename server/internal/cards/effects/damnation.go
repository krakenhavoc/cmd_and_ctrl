package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Damnation — "Destroy all creatures. They can't be regenerated."
// Functional mirror of Wrath of God in a different colour. Shares the
// wrathDestroyAllCreaturesNoRegen helper defined in wrath_of_god.go,
// which since S23 sweeps through DestroyAllMatching so the deaths are
// simultaneous and since #667 carries the CR 701.19c rider.
func init() {
	Register(Spec{
		OracleID:     "d57a8f0b-7989-4db5-8756-6f2690097252",
		Name:         "Damnation",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Completeness: CompletenessFull,
		OnResolve:    wrathDestroyAllCreaturesNoRegen,
	})
}
