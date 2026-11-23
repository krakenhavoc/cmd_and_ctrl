package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reservoir Walker — Artifact Creature — Construct {5}, 3/3:
//
//	"When this creature enters, you gain 3 life and get {E}{E}{E} (three
//	 energy counters)."
//
// ADR 0129 PR 1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7177361a-f86f-4a50-b517-229cb1eaac07",
		Name:         "Reservoir Walker",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 3},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Reservoir Walker — you gain 3 life and get {E}{E}{E}", effects.GainLifeAndGetEnergy(3, 3)),
		},
	})
}
