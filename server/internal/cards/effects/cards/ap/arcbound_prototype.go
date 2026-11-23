package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcbound Prototype — Artifact Creature — Assembly-Worker {1}{W}, 0/0
// (Modern Horizons 2, #5501):
//
//	"Modular 2 (This creature enters with two +1/+1 counters on it. When
//	 it dies, you may put its +1/+1 counters on target artifact
//	 creature.)"
//
// Arcbound Reclaimer's modular 2 without its other ability: the engine's
// keyword (#2012, game/modular.go, CR 702.43a) carries both halves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ac1003b4-1f7c-4801-9dd3-5cb7d9881af8",
		Name:            "Arcbound Prototype",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"modular 2"},
	})
}
