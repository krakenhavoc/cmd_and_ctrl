package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcbound Mouser — Artifact Creature — Cat {W}, 0/0 (Modern Horizons 2,
// #5501):
//
//	"Lifelink
//	 Modular 1 (This creature enters with a +1/+1 counter on it. When it
//	 dies, you may put its +1/+1 counters on target artifact creature.)"
//
// Arcbound Worker with lifelink: both are engine keywords (modular:
// #2012, game/modular.go, CR 702.43a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5e59199c-facc-46b4-9681-988de0cfaba2",
		Name:            "Arcbound Mouser",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink", "modular 1"},
	})
}
