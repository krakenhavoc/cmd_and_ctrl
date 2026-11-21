package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sun Sentinel — Creature — Human Soldier {1}{W}, 2/2 (Rivals of Ixalan, #5531):
//
//	"Vigilance (Attacking doesn't cause this creature to tap.)"
//
// Vigilance is an engine keyword carried as printed data: attacking
// doesn't tap it (CR 702.20b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4db00a5e-7dbe-4387-90e9-c21ec1daa448",
		Name:            "Sun Sentinel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
