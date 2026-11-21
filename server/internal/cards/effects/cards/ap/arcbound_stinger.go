package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcbound Stinger — Artifact Creature — Insect {2}, 0/0 (Darksteel,
// #5605):
//
//	"Flying
//	 Modular 1 (This creature enters with a +1/+1 counter on it. When it
//	 dies, you may put its +1/+1 counters on target artifact creature.)"
//
// Flying and modular are both engine keywords (modular: #2012,
// game/modular.go, CR 702.43a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7e6e2c37-3a06-456a-ae0f-37b7dc1a3e2b",
		Name:            "Arcbound Stinger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "modular 1"},
	})
}
