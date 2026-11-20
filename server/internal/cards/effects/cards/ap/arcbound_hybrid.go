package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcbound Hybrid — Artifact Creature — Beast {4}, 0/0 (Darksteel, #5605):
//
//	"Haste
//	 Modular 2 (This creature enters with two +1/+1 counters on it. When
//	 it dies, you may put its +1/+1 counters on target artifact
//	 creature.)"
//
// Haste and modular are both engine keywords (modular: #2012,
// game/modular.go, CR 702.43a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4bd7b380-ed4b-46a1-8211-1c4b651a52c6",
		Name:            "Arcbound Hybrid",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste", "modular 2"},
	})
}
