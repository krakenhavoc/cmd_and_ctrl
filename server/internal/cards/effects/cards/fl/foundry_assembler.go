package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Foundry Assembler — Artifact Creature — Assembly-Worker {5}, 3/3:
//
//	"Improvise (Your artifacts can help cast this spell. Each artifact
//	 you tap after you're done activating mana abilities pays for {1}.)"
//
// Improvise is Spec.TapCost (Improvise(), CR 702.126); nothing else is
// printed. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8242f4f5-c51d-42dc-a6ae-67060368fed9",
		Name:         "Foundry Assembler",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Improvise(),
	})
}
