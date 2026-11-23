package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steel Wall — Artifact Creature — Wall {1}, 0/4 (Mirrodin, #5606):
//
//	"Defender (This creature can't attack.)"
//
// Defender is a printed keyword the attack declaration honours
// (CR 702.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5ccb57e1-ca94-4b5a-8e5f-b8b5e692cfb9",
		Name:            "Steel Wall",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
	})
}
