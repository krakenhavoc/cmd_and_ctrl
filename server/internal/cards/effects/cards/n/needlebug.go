package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Needlebug — Artifact Creature — Insect {4}, 2/2 (Mirrodin, #5606):
//
//	"Flash
//	 Protection from artifacts"
//
// Flash lets it be cast any time its controller could cast an instant
// (CR 702.8a), and protection from artifacts is read by targeting,
// blocking, damage and attaching (CR 702.16a), as on Nacatl Savage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2d5fba80-c84f-4f84-84e4-9c797ac244fc",
		Name:            "Needlebug",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "protection from artifacts"},
	})
}
