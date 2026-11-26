package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dreg Reaver — {4}{B} Creature — Zombie Beast 4/3 (Shards of Alara, S82 OLD b513):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see grizzly_bears.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ecc5467d-d327-4a9c-b935-09afbea111dc",
		Name:         "Dreg Reaver",
		Completeness: effects.CompletenessFull,
	})
}
