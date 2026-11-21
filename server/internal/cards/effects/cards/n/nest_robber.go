package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nest Robber — Creature — Dinosaur {1}{R}, 2/1 (Ixalan, #5532):
//
//	"Haste"
//
// Haste is an engine keyword carried as printed data: it can attack and
// {T} the turn it arrives (CR 702.10b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b909ff27-423f-4e84-887b-7840149a85be",
		Name:            "Nest Robber",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
