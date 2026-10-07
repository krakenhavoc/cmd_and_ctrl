package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reservoir Walker — Artifact Creature — Construct {5}, 3/3:
//
//	"When this creature enters, you gain 3 life and get {E}{E}{E} (three
//	 energy counters)."
//
// ADR 0129 PR 1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7177361a-f86f-4a50-b517-229cb1eaac07",
		Name:         "Reservoir Walker",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 3},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Reservoir Walker — you gain 3 life and get {E}{E}{E}", gainLifeAndGetEnergy(3, 3)),
		},
	})
}
