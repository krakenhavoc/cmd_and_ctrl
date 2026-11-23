package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fortified Rampart — Creature — Wall {1}{W}, 0/6 (Battle for Zendikar,
// #5547):
//
//	"Defender"
//
// Defender is an engine keyword carried as printed data: it can't
// attack (CR 702.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a55202a9-bddb-40df-b93f-06678ae1e0ea",
		Name:            "Fortified Rampart",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
	})
}
